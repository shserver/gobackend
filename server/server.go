package main

import (
	"chat-service/chatpb"
	"context"
	"database/sql"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-jwt/jwt/v4"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
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
	Accounts map[string]string
	db       *sql.DB
}

const (
	sqlDriver     = "mysql"
	dbServer      = "root:ksh0917@tcp(127.0.0.1:3306)/chat-service"
	httpServer    = "0.0.0.0:8080"
	httpsServer   = "0.0.0.0:50000"
	grpcServer    = "0.0.0.0:8081"
	grpcTlsServer = "0.0.0.0:50001"
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	return string(bytes), err
}
func CheckHashPassword(password string, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func (s *server) SignUp(ctx context.Context, req *chatpb.RequestSignUp) (*chatpb.ResponseSignUp, error) {
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
		return &chatpb.ResponseSignUp{Success: false},
			status.Errorf(codes.InvalidArgument, fmt.Sprintf("Invalid Type"))
	}

	err = s.db.QueryRow("SELECT id FROM users WHERE id=?", id).Scan(&id)
	if err == sql.ErrNoRows {
		hash_pw, err := HashPassword(pw)
		if err != nil {
			log.Printf("hash password error :%v", err)
			return &chatpb.ResponseSignUp{Success: false},
				status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
		}
		_, err = s.db.Exec("INSERT INTO users VALUES (?, ?, ?, ?)", id, hash_pw, name, email)
		if err != nil {
			log.Printf("INSERT into users DB failed : %v", err)
			return &chatpb.ResponseSignUp{Success: false},
				status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
		}
		log.Printf("New Account ID :%s", id)
		return &chatpb.ResponseSignUp{Success: true}, nil
	} else if err == nil {
		log.Println("Already exist")
		return &chatpb.ResponseSignUp{Success: false},
			status.Errorf(codes.AlreadyExists, fmt.Sprintf("Alreay exists"))
	} else {
		log.Printf("Internal error: %v", err)
		return &chatpb.ResponseSignUp{Success: false},
			status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
	}
}

func (s *server) SignIn(ctx context.Context, req *chatpb.RequestSignIn) (*chatpb.ResponseSignIn, error) {
	log.Print("SignIn Request from client")
	id := req.GetId()
	pw := req.GetPw()
	hash_pw := ""

	err := s.db.QueryRow("SELECT pw FROM users WHERE id=?", id).Scan(&hash_pw)
	if err == sql.ErrNoRows {
		log.Printf("query err %v", err)
		return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Incorrect user or password"))
	} else if err == nil {
		if !CheckHashPassword(pw, hash_pw) {
			log.Printf("wrong password")
			return nil, status.Errorf(codes.NotFound, fmt.Sprintf("Incorrect user or password"))
		}
		//jwt
		claims := jwt.RegisteredClaims{
			Issuer: "sehyoung",
			// Subject:   "chat-server",
			// Audience:  []string{"chat-customer"},
			ExpiresAt: jwt.NewNumericDate(time.Unix(60, 0)),
			// ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 3)),
			// NotBefore: jwt.NewNumericDate(time.Now()),
			// IssuedAt:  jwt.NewNumericDate(time.Now()),
			// ID:
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		my_signed_key := []byte("Mup#&^Q3ZHv5ojI1Yzcr*DWiTBsbN91p")
		signed_jwt, err := token.SignedString(my_signed_key)
		if err != nil {
			log.Printf("JWT generation failed ... %v", err)
			return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal authentication error"))
		}
		fmt.Printf("signed_jwt %s", signed_jwt)

		return &chatpb.ResponseSignIn{Auth: &chatpb.Authorization{Jwt: signed_jwt}}, nil
	} else {
		log.Printf("Internal error")
		return nil, status.Errorf(codes.Internal, fmt.Sprintf("Internal query error"))
	}
}

func (s *server) Chat(stream chatpb.ChatService_ChatServer) error {
	log.Println("Chat request from client")
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.Printf("client EOF")
			return nil
		}
		if err != nil {
			log.Printf("[Chat] err : %v", err)
			return status.Errorf(codes.Canceled, fmt.Sprintf("receive error from client %v", err))
		}
		id := req.GetId()
		message := "Thank you" + req.GetMessage()

		err = stream.Send(&chatpb.ResponseMessage{Id: id, Message: message})
		if err != nil {
			return status.Errorf(codes.Canceled, fmt.Sprintf("send error to client %v", err))
		}
	}
}

func main() {
	// DB
	db, err := sql.Open(sqlDriver, dbServer)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// TLS

	// Server
	lis, err := net.Listen("tcp", grpcServer)
	if err != nil {
		log.Println("Listen error")
		panic(err)
	}

	s := grpc.NewServer()
	chatpb.RegisterChatServiceServer(s, &server{
		Accounts: make(map[string]string),
		db:       db,
	})
	go func() {
		err := s.Serve(lis)
		if err != nil {
			log.Printf("grpc server error")
		}
	}()

	mux := runtime.NewServeMux()
	// HTTPS
	cert, err := credentials.NewServerTLSFromFile("../tls/server.crt", "../tls/server.pem")
	if err != nil {
		panic(err)
	}
	opts := []grpc.DialOption{grpc.WithTransportCredentials(cert)}
	// HTTP
	// opts := []grpc.DialOption{grpc.WithInsecure()}
	err = chatpb.RegisterChatServiceHandlerFromEndpoint(context.Background(), mux, grpcServer, opts)
	if err != nil {
		panic(err)
	}
	// HTTPS
	err = http.ListenAndServeTLS(httpsServer, "../tls/server.crt", "../tls/server.pem", mux)
	// HTTP
	// http.ListenAndServe(httpServer, mux)
	if err != nil {
		log.Println("rest server error")
		panic(err)
	}
}
