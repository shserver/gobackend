package main

import (
	"log"
	"os/exec"
	"time"
)

func main() {

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
