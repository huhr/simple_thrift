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

	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()

	var transport thrift.TTransport

	transport, err := thrift.NewTSocket("10.0.0.206:9903")
	if err != nil {
		fmt.Println(err.Error())
	}

	transport = transportFactory.GetTransport(transport)
	err = transport.Open()
	if err != nil {
		fmt.Println(err.Error())
	}

	client := hello.NewUserManagerClientFactory(transport, protocolFactory)

	for true {
		fmt.Println("start call\n")
		time.Sleep(10)
		user, err := client.GetUser(1)
		if err != nil {
			fmt.Println(err.Error())
		}
		fmt.Printf("User: %+v", user)
	}
}
