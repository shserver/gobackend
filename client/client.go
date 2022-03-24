package main

import (
	"context"
	"fmt"
	"io"
	"log"
	pb "sehyoung/pb/gen"

	"google.golang.org/grpc"
)

const (
	address            = "0.0.0.0:8080"
	testServiceAddress = "0.0.0.0:8081"
	authServiceAddress = "0.0.0.0:50000"
	chatServiceAddress = "0.0.0.0:50001"
)

func chat(clnt pb.ChatServiceClient) {

	stream, err := clnt.Chat(context.Background())
	if err != nil {
		log.Fatal("chat request failed to server ...")
	}
	wait := make(chan struct{})
	go func() {
		for {
			message := ""
			fmt.Print("Enter message : ")
			fmt.Scanf("%s", &message)
			if message == "-1" {
				fmt.Println("Close chat service")
				break
			}
			stream.Send(&pb.RequestMessage{
				Message: message})
		}
		stream.CloseSend()
	}()

	go func() {
		for {
			log.Println("recv start")
			res, err := stream.Recv()
			if err == io.EOF {
				log.Println("stream Recv EOF")
				break
			} else if err != nil {
				log.Fatalf("response error from chat-server : %v", err)
				break
			}
			log.Printf("response from chat-server : %v", res)
		}
		close(wait)
	}()
	<-wait
}

func main() {
	opt := grpc.WithInsecure()
	conn, err := grpc.Dial(chatServiceAddress, opt)
	if err != nil {
		log.Fatal("connect failed ...")
	}
	defer conn.Close()

	// clnt := pb.NewChatServiceClient(conn)
	clnt := pb.NewChatServiceClient(conn)

	chat(clnt)
}
