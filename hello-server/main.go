package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"hello-server/interceptor"
	"hello-server/service"
)

// hello server

func main() {
	// 监听本地端口8972
	lis, err := net.Listen("tcp", ":8972")
	if err != nil {
		fmt.Printf("Failed to listen: %v", err)
		return
	}
	creds, _ := credentials.NewServerTLSFromFile("../server.crt", "../server.key")
	s := grpc.NewServer(
		grpc.Creds(creds),
		// 服务端注册拦截器
		grpc.UnaryInterceptor(interceptor.UnaryInterceptor),
		grpc.StreamInterceptor(interceptor.StreamInterceptor),
	) // 创建 gRPC 服务器
	service.RegisterGreeterServer(s) // 在 gRPC 服务端注册服务
	// 启动服务
	err = s.Serve(lis)
	if err != nil {
		fmt.Printf("Failed to serve: %v", err)
		return
	}
}
