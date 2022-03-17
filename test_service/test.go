package main

import (
	"context"
	"log"
	"net"
	pb "sehyoung/pb/gen"

	"google.golang.org/grpc"
)

type server struct {
}

const (
	TLS           = true
	sqlDriver     = "mysql"
	dbServer      = "root:ksh0917@tcp(127.0.0.1:3306)/chat-service"
	httpServer    = "0.0.0.0:8080"
	httpsServer   = "0.0.0.0:50000"
	grpcServer    = "0.0.0.0:8081"
	grpcTlsServer = "0.0.0.0:50001"
)

func (s *server) Hello(ctx context.Context, req *pb.TestMessage) (*pb.TestMessage, error) {
	log.Printf("Hello from client: %s", req.GetMsg())
	return &pb.TestMessage{Msg: "Welcome !!!"}, nil
}

func main() {
	// Server
	lis, err := net.Listen("tcp", grpcServer)
	if err != nil {
		log.Println("Listen error")
		panic(err)
	}

	s := grpc.NewServer()

	pb.RegisterTestServiceServer(s, &server{})
	err = s.Serve(lis)
	if err != nil {
		log.Printf("grpc server error")
	}
}
