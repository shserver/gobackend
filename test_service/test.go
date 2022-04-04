package main

import (
	"context"
	"net"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"

	"github.com/shserver/gopackage/shlog"
	"google.golang.org/grpc"
)

type server struct {
}

const (
	address = "0.0.0.0:8081"
)

func (s *server) Hello(ctx context.Context, req *pb.TestMessage) (*pb.TestMessage, error) {
	shlog.Logf("INFO", "Hello from client: %s", req.GetMsg())

	return &pb.TestMessage{Msg: "Welcome !!!"}, nil
}

func main() {
	shlog.InitLogger("")
	// Server
	lis, err := net.Listen("tcp", address)
	if err != nil {
		shlog.Logf("FATAL", "Listen error: %v", err)
	}
	opts := []grpc.ServerOption{
		middleware.UnaryServer(),
	}
	s := grpc.NewServer(opts...)

	pb.RegisterTestServiceServer(s, &server{})
	shlog.Logf("INFO", "test service start...")
	err = s.Serve(lis)
	if err != nil {
		shlog.Logf("FATAL", "grpc server error: %v", err)
	}
}
