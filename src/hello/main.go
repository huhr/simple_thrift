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
	// 阻塞主进程
	for true {
		time.Sleep(100 * time.Second)
	}
}
