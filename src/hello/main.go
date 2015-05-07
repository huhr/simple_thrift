package main

import (
	"fmt"
	"time"

	"hello/client"
	"hello/server"
)

func main() {
	server := server.NewServer()
	go server.Run()
	time.Sleep(5)
	fmt.Println("sleep a while")
	client := client.NewClient()
	go client.Run()
	for true {
		time.Sleep(10 * time.Second)
	}
}
