package main

import (
	"errors"
	"fmt"
	"log"
	"os/exec"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func main() {
	err := errors.New("test")
	test1 := status.Errorf(codes.Unauthenticated, "Invalid authorization token, %v", err)
	test2 := status.Errorf(codes.Unauthenticated, fmt.Sprintf("Invalid authorization token, %v", err))

	log.Println(test1)
	log.Println(test2)

}

func terminalExec() {
	start_time := time.Now()
	for i := 0; i < 10; i++ {
		elapsed_time := time.Since(start_time)
		cmd := exec.Command("go", "run", "../client/client.go", elapsed_time.String())
		// cmd.Dir = "../client/"
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}
		time.Sleep(1 * time.Microsecond)
	}
}
