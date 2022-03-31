package main

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"
	"strings"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/shserver/gopackage/shemail"
	"github.com/shserver/gopackage/shlog"
	"github.com/shserver/gopackage/shorm"
	"github.com/shserver/gopackage/shvalidator"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// account table definition
// If the number of DB rows is 0, the role will be "admin" from SignUp funcion.
// email(user id), pw hashing, role: admin, user ...
type Account struct {
	id          int    `shorm:"primary key;auto_increment"`
	email       string `shorm:"varchar(40);unique;not null"`
	password    string `shorm:"varchar(64);not null"`
	phoneNumber string `shorm:"varchar(64);not null"`
	name        string `shorm:"varchar(20);not null"`
	role        string `shorm:"varchar(10);not null"`
}

type server struct {
	userDB   *sql.DB    //id email pw name phonenumber role
	redisCon redis.Conn //id refreshToken
}

const (
	serverAddress = "0.0.0.0:50000"
	sqlDriver     = "mysql"
	tableName     = "account"
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
	shlog.Logf("INFO", "SignUp Request from client")
	email := req.GetEmail()
	password := req.GetPassword()
	name := req.GetName()
	phoneNumber := req.GetPhoneNumber()

	// Check validation
	validate := validator.New()
	validate.RegisterValidation("phonenumber", shvalidator.IsValidPhoneNumber)
	validate.RegisterValidation("password", shvalidator.IsValidPassword)

	// why to print number log in this funcion.
	err := validate.Struct(&shvalidator.AccountValidator{
		Email:       email,
		Password:    password,
		Name:        name,
		PhoneNumber: phoneNumber,
	})
	if err != nil {
		shlog.Logf("INFO", "Validation failed: %v", err)
		return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_INVALID_FORMAT},
			status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid Type"))
	}

	// The user would be admin if DB data's count is 0.
	count := 0
	err = s.userDB.QueryRow(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
	if err != nil {
		return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_FAIL}, status.Errorf(codes.Internal, "Internal query error: %w", err)
	}
	role := "user"
	if count == 0 {
		role = "admin"
	}

	// Check duplicated ID
	id := ""
	err = s.userDB.QueryRow(fmt.Sprintf("SELECT id FROM %s WHERE email=?", tableName), email).Scan(&id)
	if err == sql.ErrNoRows {
		hash_pw, err := HashPassword(password)
		if err != nil {
			shlog.Logf("INFO", "hash password error: %v", err)
			return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_FAIL},
				status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
		}
		// email authentication
		err = shemail.SendEmail("kimsh9688@naver.com", "인증 번호 입력요청", "테스트")
		if err != nil {
			shlog.Logf("INFO", "email authentication: %v", err)
		}

		// Insert user info into DB
		_, err = s.userDB.Exec(fmt.Sprintf("INSERT INTO %s (email, password, phoneNumber, name, role) VALUES (?, ?, ?, ?, ?)", tableName), email, hash_pw, phoneNumber, name, role)
		if err != nil {
			shlog.Logf("ERROR", "INSERT into %s DB failed: %v", tableName, err)
			return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_FAIL}, status.Errorf(codes.Internal, "Internal query error")
		}
		shlog.Logf("INFO", "New Account ID(email): %s", email)
		return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_SUCCESS}, nil
	} else if err == nil {
		return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_FAIL}, status.Errorf(codes.AlreadyExists, "Alreay exists")
	} else {
		return &pb.ResponseSignUp{Validation: pb.ResponseSignUp_FAIL}, status.Errorf(codes.Internal, "Internal query error: %w", err)
	}
}

func (s *server) SignIn(ctx context.Context, req *pb.RequestSignIn) (*pb.ResponseSignIn, error) {
	shlog.Logf("INFO", "SignIn Request from client")
	email := req.GetEmail()
	password := req.GetPassword()
	hash_pw := ""

	err := s.userDB.QueryRow(fmt.Sprintf("SELECT password FROM %s WHERE email=?", tableName), email).Scan(&hash_pw)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "Incorrect user or password")
	} else if err == nil {
		if !CheckHashPassword(password, hash_pw) {
			return nil, status.Errorf(codes.NotFound, "Incorrect user or password")
		}
		signedJWT, err := middleware.CreateJWT(email, middleware.TokenDuration("access"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}
		refreshSignedJWT, err := middleware.CreateJWT(email, middleware.TokenDuration("refresh"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}

		_, err = s.redisCon.Do("SET", email, refreshSignedJWT)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Internal authentication error")
		}
		return &pb.ResponseSignIn{Token: signedJWT, RefreshToken: refreshSignedJWT}, nil
	} else {
		return nil, status.Errorf(codes.Internal, "Internal query error")
	}
}

func (s *server) SignOut(ctx context.Context, req *pb.RequestSignOut) (*pb.ResponseSignOut, error) {
	shlog.Logf("INFO", "SignOut Request from client")
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
	shlog.Logf("INFO", "%s's token r(%d) was deleted", aud, r)
	if r == 0 {
		return nil, status.Errorf(codes.Unauthenticated, "Already deleted token")
	}

	return &pb.ResponseSignOut{}, nil
}

func (s *server) RefreshToken(ctx context.Context, req *pb.RequestRefreshToken) (*pb.ResponseRefreshToken, error) {
	shlog.Logf("INFO", "RefreshToken Request from client")
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
		shlog.Logf("INFO", "Auth Fail : %v", err)
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

func (s *server) DeleteAccount(ctx context.Context, req *pb.RequestDeleteAccount) (*pb.ResponseDeleteAccount, error) {
	shlog.Logf("INFO", "DeleteAccount Request from client")
	password := req.GetPassword()
	email, err := middleware.AudInfoFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Check whether email is exist and password is correct.
	hashPassword := ""
	err = s.userDB.QueryRow(fmt.Sprintf("SELECT password FROM %s WHERE email=?", tableName), email).Scan(&hashPassword)
	if err == sql.ErrNoRows {
		return nil, status.Errorf(codes.NotFound, "Incorrect user or password")
	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal query error")
	}

	if !CheckHashPassword(password, hashPassword) {
		return nil, status.Errorf(codes.InvalidArgument, "Wrong password")
	}

	// Delete refresh token
	r, err := redis.Int(s.redisCon.Do("DEL", email))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal error: %w", err)
	}
	shlog.Logf("INFO", "%s's token r(%d) has been deleted", email, r)

	// Delete user info from DB
	_, err = s.userDB.Exec(fmt.Sprintf("DELETE FROM %s WHERE email=?", tableName), email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Internal error: %w", err)
	}
	return &pb.ResponseDeleteAccount{}, nil
}
func (s *server) FindID(ctx context.Context, req *pb.RequestFindID) (*pb.ResponseFindID, error) {

	return nil, nil
}
func (s *server) FindPW(ctx context.Context, req *pb.RequestFindPW) (*pb.ResponseFindPW, error) {

	return nil, nil
}

func main() {
	shlog.InitLogger("")

	err := godotenv.Load("../.env")
	if err != nil {
		shlog.Logf("FATAL", "Load env error: %v", err)
	}

	// user DB server
	db, err := sql.Open(sqlDriver, os.Getenv("USER_DB_SERVER"))
	if err != nil {
		shlog.Logf("FATAL", "can't connect to gobackend server: %v", err)
	}
	defer db.Close()

	err = shorm.CreateTable(db, &Account{})
	// _, err = db.Exec(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s(id int primary key auto_increment, role VARCHAR(10) not null, user_id VARCHAR(32) unique not null, pw VARCHAR(64) not null, name VARCHAR(20) not null, email VARCHAR(40) not null)", tableName))
	if err != nil {
		shlog.Logf("FATAL", "can't create %s table: %v", tableName, err)
	}

	// email
	shemail.SetSmtpPassword()

	// token DB server
	c, err := redis.Dial("tcp", ":6379")
	if err != nil {
		shlog.Logf("FATAL", "can't connect to token server: %v", err)
	}
	defer c.Close()

	// grpc server
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		shlog.Logf("FATAL", "can't open auth server: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{
		userDB:   db,
		redisCon: c,
	})
	shlog.Logf("INFO", "auth service start...")
	err = s.Serve(lis)
	if err != nil {
		shlog.Logf("FATAL", "auth server error: ", err)
	}
}
