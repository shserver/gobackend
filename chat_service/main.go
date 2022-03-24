package main

import (
	"io"
	"log"
	"net"
	pb "sehyoung/pb/gen"
	"sehyoung/server/middleware"

	_ "github.com/go-sql-driver/mysql"
	"google.golang.org/grpc"
)

type server struct {
}

const (
	serverAddress = "0.0.0.0:50001"
)

func (s *server) Chat(stream pb.ChatService_ChatServer) error {
	log.Println("Chat request from client")
	wait := make(chan struct{})
	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				log.Printf("client EOF")
				break
			}
			if err != nil {
				log.Printf("[Chat] err : %v", err)
				break
			}
			log.Print("request from client: ", req.GetMessage())
		}
	}()

	go func() {
		send(stream)
		close(wait)
	}()
	<-wait
	return nil
}

func send(stream pb.ChatService_ChatServer) {
	for {
		select {
		case <-stream.Context().Done():
			return
		default:
			err := stream.Send(&pb.ResponseMessage{Message: message})
			if err != nil {
				log.Println("send err!!!!!!")
				return
			}
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
