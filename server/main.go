package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"
	"sehyoung/server/ws"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"github.com/shserver/gopackage/shlog"
	"google.golang.org/grpc"
)

const (
	serverIP             = ""
	serverPort           = 8080
	serverTLS            = false
	serverCRT            = "../tls/server.crt"
	serverKey            = "../tls/server.pem"
	testServiceAddress   = "0.0.0.0:8081"
	authServiceAddress   = "0.0.0.0:50000"
	chatServiceAddress   = "0.0.0.0:50001"
	publicServiceAddress = "0.0.0.0:50002"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func runServer() error {
	address := fmt.Sprintf("%s:%d", serverIP, serverPort)
	lis, err := net.Listen("tcp", address)
	if err != nil {
		shlog.Logf("FATAL", "can't run server: %v", err)
	}

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(middleware.UnaryClient()),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Websocket
	err = ws.RegisterWebsocketMux(mux)
	if err != nil {
		shlog.Logf("FATAL", "can't register websocket mux: %v", err)
	}
	// register service
	err = pb.RegisterPublicServiceHandlerFromEndpoint(ctx, mux, publicServiceAddress, opts)
	if err != nil {
		shlog.Logf("FATAL", "can't register public service: %v", err)
	}
	err = pb.RegisterTestServiceHandlerFromEndpoint(ctx, mux, testServiceAddress, opts)
	if err != nil {
		shlog.Logf("FATAL", "can't register test service: ", err)
	}
	err = pb.RegisterAuthServiceHandlerFromEndpoint(ctx, mux, authServiceAddress, opts)
	if err != nil {
		shlog.Logf("FATAL", "can't register auth service: ", err)
	}

	shlog.Logf("INFO", "server start..")
	if serverTLS {
		return http.ServeTLS(lis, mux, serverCRT, serverKey)
	} else {
		return http.Serve(lis, mux)
	}
}

func main() {
	// port := flag.Int("port", serverPort, "server port")
	// tls := flag.Bool("tls", serverTLS, "TLS flag")
	// flag.Parse()
	shlog.InitLogger("")

	err := godotenv.Load("../.env")
	if err != nil {
		shlog.Logf("FATAL", "Load env error: %v", err)
	}

	middleware.LoadSecretKey(os.Getenv("JWT_SECRET_KEY"))

	err = runServer()
	if err != nil {
		shlog.Logf("FATAL", "runServer error: %v", err)
	}
}
