package main

import (
	"fmt"
	"io"
	"log"
	"net"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type server struct {
}

const (
	serverAddress = "0.0.0.0:50001"
)

func (s *server) Chat(stream pb.ChatService_ChatServer) error {
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
		message := "Thank you" + req.GetMessage()

		err = stream.Send(&pb.ResponseMessage{Message: message})
		if err != nil {
			return status.Errorf(codes.Canceled, fmt.Sprintf("send error to client %v", err))
		}
	}
}

func main() {
	lis, err := net.Listen("tcp", serverAddress)
	if err != nil {
		log.Println("Listen error")
		panic(err)
	}
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(middleware.UnaryServer()),
	}
	s := grpc.NewServer(opts...)

	pb.RegisterChatServiceServer(s, &server{})
	log.Printf("chat service start...")
	err = s.Serve(lis)
	if err != nil {
		log.Printf("grpc server error")
	}
}
