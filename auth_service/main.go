package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"os"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"
	"strings"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// validator is designed to be thread-safe and used as a singleton instance.
type Account struct {
	Id    string `validate:"required,max=15,min=5"`
	Pw    string `validate:"required,max=15,min=8"`
	Name  string `validate:"required"`
	Email string `validate:"required,email"`
}

type server struct {
	userDB   *sql.DB    //id pw name email
	redisCon redis.Conn //id refreshToken
}

var (
	sqlDriver     = "mysql"
	serverAddress = "0.0.0.0:50000"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
func CheckHashPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *server) SignUp(ctx context.Context, req *pb.RequestSignUp) (*pb.ResponseSignUp, error) {
	log.Print("SignUp Request from client")
	id := req.GetId()
	pw := req.GetPw()
	name := req.GetName()
	email := req.GetEmail()

	validate := validator.New()
	err := validate.Struct(&Account{
		Id:    id,
		Pw:    pw,
		Name:  name,
		Email: email,
	})
	if err != nil {
		log.Printf("Validation failed : %v", err)
		return &pb.ResponseSignUp{Success: false},
			status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid Type"))
	}

	err = s.userDB.QueryRow("SELECT id FROM users WHERE id=?", id).Scan(&id)
	if err == sql.ErrNoRows {
		hash_pw, err := HashPassword(pw)
		if err != nil {
			log.Printf("hash password error :%v", err)
			return &pb.ResponseSignUp{Success: false},
				status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
		}
		_, err = s.userDB.Exec("INSERT INTO users VALUES (?, ?, ?, ?)", id, hash_pw, name, email)
		if err != nil {
			log.Printf("INSERT into users DB failed : %v", err)
			return &pb.ResponseSignUp{Success: false}, status.Errorf(codes.Internal, "Internal query error")
		}
		log.Printf("New Account ID :%s", id)
		return &pb.ResponseSignUp{Success: true}, nil
	} else if err == nil {
		return &pb.ResponseSignUp{Success: false}, status.Errorf(codes.AlreadyExists, "Alreay exists")
	} else {
		return &pb.ResponseSignUp{Success: false}, status.Errorf(codes.Internal, "Internal query error: %w", err)
	}
}

func (s *server) SignIn(ctx context.Context, req *pb.RequestSignIn) (*pb.ResponseSignIn, error) {
	log.Print("SignIn Request from client")
	id := req.GetId()
	pw := req.GetPw()
	hash_pw := ""

	err := s.userDB.QueryRow("SELECT pw FROM users WHERE id=?", id).Scan(&hash_pw)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "Incorrect user or password")
	} else if err == nil {
		if !CheckHashPassword(pw, hash_pw) {
			return nil, status.Errorf(codes.NotFound, "Incorrect user or password")
		}
		signedJWT, err := middleware.CreateJWT(id, middleware.TokenDuration("access"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}
		refreshSignedJWT, err := middleware.CreateJWT(id, middleware.TokenDuration("refresh"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}

		_, err = s.redisCon.Do("SET", id, refreshSignedJWT)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}
		return &pb.ResponseSignIn{Token: signedJWT, RefreshToken: refreshSignedJWT}, nil
	} else {
		return nil, status.Errorf(codes.Internal, "Internal query error")
	}
}

func (s *server) SignOut(ctx context.Context, req *pb.RequestSignOut) (*pb.ResponseSignOut, error) {
	log.Print("SignOut Request from client")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "No metadata")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "No authorization token")
	}

	claims, err := middleware.VerifyJWT(values[0])
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization token: %w", err)
	}

	aud := claims.Audience[0]
	r, err := redis.Int(s.redisCon.Do("DEL", aud))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal authentication error: %w", err)
	}

	log.Printf("%s's token r(%d) was deleted", aud, r)

	return &pb.ResponseSignOut{}, nil
}

func (s *server) RefreshToken(ctx context.Context, req *pb.RequestRefreshToken) (*pb.ResponseRefreshToken, error) {
	log.Print("RefreshToken Request from client")
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "No metadata")
	}
	values := md["authorization"]
	if len(values) == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "No authorization token")
	}

	claims, err := middleware.VerifyJWT(values[0])
	if err != nil {
		log.Printf("Auth Fail : %v", err)
		return nil, status.Errorf(codes.Unauthenticated, "Invalid authorization token, %w", err)
	}

	aud := claims.Audience[0]
	r, err := redis.String(s.redisCon.Do("GET", aud))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err: ", err)
	}

	if strings.Contains(values[0], r) == false {
		return nil, status.Errorf(codes.Unauthenticated, "Non matching token")
	}

	signedJWT, err := middleware.CreateJWT(aud, middleware.TokenDuration("access"))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal authentication error: %w", err)
	}

	refreshSignedJWT := ""
	if middleware.RefreshTokenReissue(claims.ExpiresAt.Time) {
		refreshSignedJWT, err = middleware.CreateJWT(aud, middleware.TokenDuration("refresh"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error: %w", err)
		}
		_, err = s.redisCon.Do("SET", aud, refreshSignedJWT)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error: %w", err)
		}
	}

	return &pb.ResponseRefreshToken{Token: signedJWT, RefreshToken: refreshSignedJWT}, nil
}

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatal("Load env error: ", err)
	}

	// user DB server
	db, err := sql.Open(sqlDriver, os.Getenv("USER_DB_SERVER"))
	if err != nil {
		log.Fatal("can't connect to user server: ", err)
	}
	defer db.Close()

	// token DB server
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		log.Fatal("can't connect to token server: ", err)
	}
	defer c.Close()

	// grpc server
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Fatal("can't open auth server: ", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{
		userDB:   db,
		redisCon: c,
	})
	log.Printf("auth service start...")
	err = s.Serve(lis)
	if err != nil {
		log.Fatal("auth server error: ", err)
	}
}
