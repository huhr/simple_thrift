// client端程序
package client

import (
	"fmt"
	"time"

	"gen/hello"
	"thrift"
)

type Client struct {
}

func NewClient() *Client {
	return &Client{}
}

func (c *Client) Run() {
	fmt.Println("start client")
	// 这里transport和protocol的选择要与server端匹配
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	transport, err := thrift.NewTSocket("10.0.0.206:9903")
	if err != nil {
		fmt.Println(err.Error())
	}
	err = transport.Open()
	if err != nil {
		fmt.Println(err.Error())
	}
	// client的参数较为简单
	client := hello.NewUserManagerClientFactory(transport, protocolFactory)

	for true {
		fmt.Println("start call")
		user, err := client.GetUser(1)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("User: %+v \n", user)
		time.Sleep(10 * time.Second)
	}
}
