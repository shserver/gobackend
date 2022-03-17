package main

import (
	"context"
	"log"
	"sehyoung-world/chatpb"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	TLS         = false
	grpcPort    = "8081"
	grpcTlsPort = "50001"
)

func hello(clnt chatpb.TestServiceClient) {
	_, err := clnt.Hello(context.Background(), &chatpb.TestMessage{Msg: "test hello"})
	if err != nil {
		log.Fatalf("sign in request failed to server ... %v", err)
	}
}

// func signUp(clnt chatpb.ChatServiceClient) {
// 	fmt.Println("singUp mode ...")
// 	id, pw, name, email := "", "", "", ""

// 	fmt.Print("id pw name email 입력")
// 	fmt.Scanf("%s %s %s %s", &id, &pw, &name, &email)

// 	res, err := clnt.SignUp(context.Background(), &chatpb.RequestSignUp{Id: id, Pw: pw, Name: name, Email: email})
// 	if err != nil {
// 		log.Fatalf("sign in request failed to server ... %v", err)
// 	}
// 	if res.GetSuccess() {
// 		log.Printf("signup success")
// 	} else {
// 		log.Printf("signup fail")
// 	}
// }

// func signIn(clnt chatpb.ChatServiceClient) string {
// 	fmt.Println("singIn mode ...")
// 	id, pw := "", ""

// 	fmt.Print("id pw 입력")
// 	fmt.Scanf("%s %s", &id, &pw)

// 	ctx := context.Background()
// 	res, err := clnt.SignIn(ctx, &chatpb.RequestSignIn{Id: id, Pw: pw})
// 	if err != nil {
// 		log.Printf("sign in request failed to server ... %v", err)
// 		return ""
// 	}
// 	chat_jwt := res.GetAuth().GetJwt()
// 	log.Println("JWT from chat-server : ", chat_jwt)
// 	return chat_jwt
// }

// func chat(clnt chatpb.ChatServiceClient) {

// 	stream, err := clnt.Chat(context.Background())
// 	if err != nil {
// 		log.Fatal("chat request failed to server ...")
// 	}
// 	wait := make(chan struct{})
// 	go func() {
// 		for {
// 			message := ""
// 			fmt.Print("Enter message : ")
// 			fmt.Scanf("%s", &message)
// 			if message == "-1" {
// 				fmt.Println("Close chat service")
// 				break
// 			}

// 			// for _, message := range messages {
// 			stream.Send(&chatpb.RequestMessage{
// 				Auth:    &chatpb.Authorization{Jwt: "temp"},
// 				Id:      "temp",
// 				Message: message})
// 			// }
// 		}
// 		stream.CloseSend()
// 	}()

// 	go func() {
// 		for {
// 			res, err := stream.Recv()
// 			if err == io.EOF {
// 				log.Println("stream Recv EOF")
// 				break
// 			} else if err != nil {
// 				log.Fatalf("response error from chat-server : %v", err)
// 				break
// 			}
// 			log.Printf("response from chat-server : %v", res)
// 		}
// 		close(wait)
// 	}()
// 	<-wait
// }

func main() {
	port := grpcPort
	opt := grpc.WithInsecure()
	if TLS {
		port = grpcTlsPort
		crd, err := credentials.NewClientTLSFromFile("../tls/ca.crt", "")
		if err != nil {
			log.Fatalf("certificate file read error: %v", err)
		}
		opt = grpc.WithTransportCredentials(crd)
	}
	conn, err := grpc.Dial("localhost:"+port, opt)
	if err != nil {
		log.Fatal("connect failed ...")
	}
	defer conn.Close()

	// clnt := chatpb.NewChatServiceClient(conn)
	clnt := chatpb.NewTestServiceClient(conn)

	// fmt.Print("select mode 1(SighUp) 2(SignIn) 3(Chat)")
	// sel := 0
	// fmt.Scanf("%d", &sel)
	// if sel == 1 {
	// 	signUp(clnt)
	// } else if sel == 2 {
	// 	token := signIn(clnt)
	// 	fmt.Printf("token from server : %s", token)
	// } else if sel == 3 {
	// 	chat(clnt)
	// } else if sel == 4 {
	hello(clnt)
	// }

	// chat(clnt, id)
}
