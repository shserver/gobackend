package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	pb "sehyoung/pb/gen"

	_ "github.com/go-sql-driver/mysql"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

const (
	serverIP           = ""
	serverPort         = 8080
	serverTLS          = true
	serverCRT          = "../tls/server.crt"
	serverPEM          = "../tls/server.pem"
	testServiceAddress = "0.0.0.0:8081"
	authServiceAddress = "0.0.0.0:50000"
	chatServiceAddress = "0.0.0.0:50001"
)

func runServer() error {
	address := fmt.Sprintf("%s:%d", serverIP, serverPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal("can't run server: ", err)
	}

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// register service
	err = pb.RegisterTestServiceHandlerFromEndpoint(ctx, mux, testServiceAddress, opts)
	if err != nil {
		log.Fatal("can't register test service")
	}
	err = pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, authServiceAddress, opts)
	if err != nil {
		log.Fatal("can't register auth service")
	}
	err = pb.RegisterChatServiceHandlerFromEndpoint(ctx, mux, chatServiceAddress, opts)
	if err != nil {
		log.Fatal("can't register chat service")
	}
	log.Println("server start")
	if serverTLS {
		return http.ServeTLS(lis, mux, serverCRT, serverPEM)
	} else {
		return http.Serve(lis, mux)
	}
}

func main() {
	// port := flag.Int("port", serverPort, "server port")
	// tls := flag.Bool("tls", serverTLS, "TLS flag")
	// flag.Parse()

	// jwtManager := service.NewJWTManager(secretKey, tokenDuration)

	err := runServer()

	if err != nil {
		log.Fatal("runServer error: ", err)
	}
}
