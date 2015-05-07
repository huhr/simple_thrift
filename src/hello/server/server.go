package server

import (
	"fmt"
	"gen/hello"
	"thrift"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Run() {
	fmt.Println("start server")
	// 使用默认的transport，不对数据流进行额外的操作
	transportFactory := thrift.NewTTransportFactory()
	// 选择序列化方式
	protocolFactory := thrift.NewTBinaryProtocolFactoryDefault()
	// 初始化server_transport
	transport, err := thrift.NewTServerSocket("10.0.0.206:9903")
	if err != nil {
		fmt.Println(err.Error())
	}
	// 接口的实现类
	hanlder := NewHandler()
	// processr类，负责函数的调度，thrift生成
	processor := hello.NewUserManagerProcessor(hanlder)
	// 服务类
	server := thrift.NewTSimpleServer4(processor, transport, transportFactory, protocolFactory)
	// 这里就阻塞住了
	err = server.Serve()
	if err != nil {
		fmt.Println(err.Error())
	}
}
