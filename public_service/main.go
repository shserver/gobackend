package main

import (
	"fmt"
	"io"
	"net"
	pb "sehyoung/pb/gen"

	"github.com/shserver/gopackage/shlog"
	"google.golang.org/grpc"
)

type server struct {
}

const (
	address = ":50002"
)

func (s *server) Counsel(stream pb.PublicService_CounselServer) error {
	shlog.Logf("INFO", "Counsel Request from client")

	wait := make(chan struct{})
	go func() {
		for {
			req, err := stream.Recv()
			if err == io.EOF {
				shlog.Logf("INFO", "Counsel recv EOF...")
				break
			} else if err != nil {
				shlog.Logf("ERROR", "Counsel recv ERROR: %v", err)
				break
			}
			message := req.GetMessage()
			shlog.Logf("INFO", "Message from client: %s", message)
		}
	}()
	go func() {
		sendCounsel(stream)
		close(wait)
	}()
	<-wait
	return nil
}

func sendCounsel(stream pb.PublicService_CounselServer) {
	for {
		shlog.Logf("INFO", "Enter Message to client: ")
		message := ""
		fmt.Scanf("%s", &message)

		select {
		case <-stream.Context().Done():
			shlog.Logf("INFO", "Counsel stream Done")
			return
		default:
			err := stream.Send(&pb.ResponseCounsel{Message: message})
			if err != nil {
				shlog.Logf("ERROR", "Counsel Send ERROR: %v", err)
				return
			}
		}
	}
}

func main() {
	shlog.InitLogger("")

	lis, err := net.Listen("tcp", address)
	if err != nil {
		shlog.Logf("FATAL", "Listen error: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterPublicServiceServer(s, &server{})
	shlog.Logf("INFO", "public service start...")
	err = s.Serve(lis)
	if err != nil {
		shlog.Logf("FATAL", "public grpc server: %v", err)
	}
}
