package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"

	"github.com/go-playground/validator"
	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
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
	userDB *sql.DB
}

const (
	sqlDriver     = "mysql"
	dbServer      = "root:ksh0917@tcp(127.0.0.1:3306)/chat-service"
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
			return &pb.ResponseSignUp{Success: false},
				status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
		}
		log.Printf("New Account ID :%s", id)
		return &pb.ResponseSignUp{Success: true}, nil
	} else if err == nil {
		log.Println("Already exist")
		return &pb.ResponseSignUp{Success: false},
			status.Errorf(codes.AlreadyExists, fmt.Sprintf("Alreay exists"))
	} else {
		log.Printf("Internal error: %v", err)
		return &pb.ResponseSignUp{Success: false},
			status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
	}
}

func (s *server) SignIn(ctx context.Context, req *pb.RequestSignIn) (*pb.ResponseSignIn, error) {
	log.Print("SignIn Request from client")
	id := req.GetId()
	pw := req.GetPw()
	hash_pw := ""

	err := s.userDB.QueryRow("SELECT pw FROM users WHERE id=?", id).Scan(&hash_pw)
	if err == sql.ErrNoRows {
		log.Printf("query err %v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Incorrect user or password"))
	} else if err == nil {
		if !CheckHashPassword(pw, hash_pw) {
			log.Printf("wrong password")
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Incorrect user or password"))
		}
		signedJWT, err := middleware.CreateJWT(id)
		if err != nil {
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal authentication error"))
		}
		fmt.Printf("signed_jwt %s", signedJWT)
		return &pb.ResponseSignIn{Auth: &pb.Authorization{Jwt: signedJWT}}, nil
	} else {
		log.Printf("Internal error")
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
	}
}

func main() {
	// auth DB server
	db, err := sql.Open(sqlDriver, dbServer)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	// grpc server
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Println("Listen error")
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &server{
		userDB: db,
	})
	err = s.Serve(lis)
	if err != nil {
		log.Printf("grpc server error")
	}
}
