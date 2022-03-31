package main

import (
	"context"
	"log"
	"net"
	pb "sehyoung/pb/gen"
	"time"

	"google.golang.org/grpc"
)

type server struct {
}

const (
	address = "0.0.0.0:8081"
)

func (s *server) Hello(ctx context.Context, req *pb.TestMessage) (*pb.TestMessage, error) {
	log.Printf("Hello from client: %s", req.GetMsg())

	// md, ok := metadata.FromIncomingContext(ctx)
	// if !ok {
	// 	return nil, status.Errorf(codes.Unauthenticated, "No metadata")
	// }

	// values := md["authorization"]
	// log.Println("test value : ", values)
	for i := 0; i < 15; i++ {
		log.Println(req.GetMsg(), ": ", i)
		time.Sleep(time.Second)
	}

	return &pb.TestMessage{Msg: "Welcome !!!"}, nil
}

func main() {
	// Serverㅜ
	lis, err := net.Listen("tcp", address)
	if err != nil {
		log.Println("Listen error")
		panic(err)
	}
	opts := []grpc.ServerOption{
		// grpc.UnaryInterceptor(middleware.UnaryServer()),
	}
	s := grpc.NewServer(opts...)

	pb.RegisterTestServiceServer(s, &server{})
	log.Printf("test service start...")
	err = s.Serve(lis)
	if err != nil {
		log.Printf("grpc server error")
	}
}
