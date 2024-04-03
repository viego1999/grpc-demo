package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/status"
	"hello-client/interceptor"
	"hello-client/pb"
)

// hello_client
const (
	defaultName = "world"
)

var (
	addr = flag.String("addr", "127.0.0.1:8972", "the address to connect to")
	name = flag.String("name", defaultName, "Name to greet")
)

func execSayHello(c pb.GreeterClient) {
	// 执行 rpc 调用并打印收到的响应数据
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	rsp, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		s := status.Convert(err)
		for _, d := range s.Details() {
			switch info := d.(type) {
			case *errdetails.QuotaFailure:
				fmt.Printf("Quota failure: %s\n", info)
			default:
				fmt.Printf("Unexcepted type: %s\n", info)
			}
		}
		log.Fatalf("could not greet: %v", err)
	}
	// 获得 RPC 响应
	log.Printf("Greeting: %s", rsp.GetMessage())
}

func execLotsOfReplies(c pb.GreeterClient) {
	// server 端流式 RPC
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	stream, err := c.LotsOfReplies(ctx, &pb.HelloRequest{Name: *name})
	if err != nil {
		log.Fatalf("c.LotsOfReplies failed, err: %v", err)
	}
	for {
		// 接受服务端返回的流式数据，当收到io.EOF或错误时退出
		rsp, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("c.LotsOfReplies failed, err: %v", err)
		}
		log.Printf("got reply: %q\n", rsp.GetMessage())
	}
}

func execLotsOfGreetings(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	// 客户端流式 RPC
	stream, err := c.LotsOfGreetings(ctx)
	if err != nil {
		log.Fatalf("c.LotsOfGreetings failed, err: %v", err)
	}
	names := []string{"七米", "q1mi", "娜扎"}
	for _, name := range names {
		// 发送流式数据
		err := stream.Send(&pb.HelloRequest{Name: name})
		if err != nil {
			log.Fatalf("c.LotsOfGreetings stream.Send(%v) failed, err: %v", name, err)
		}
	}

	rsp, err := stream.CloseAndRecv()
	if err != nil {
		log.Fatalf("c.LotsOfGreetings failed: %v", err)
	}
	log.Printf("got reply: %v", rsp.GetMessage())
}

func execBidiHello(c pb.GreeterClient) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	// 双向流式
	stream, err := c.BidiHello(ctx)
	if err != nil {
		log.Fatalf("c.BidiHello failed, err: %v", err)
	}
	waitc := make(chan struct{})
	go func() {
		for {
			// 接收服务端返回的响应
			rsp, err := stream.Recv()
			if err == io.EOF {
				// read done.
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("c.BidiHello stream.Recv() failed, err: %v", err)
			}
			fmt.Printf("AI: %s\n", rsp.GetMessage())
		}
	}()
	// 从标准输入获取用户输入
	reader := bufio.NewReader(os.Stdin)
	for {
		cmd, _ := reader.ReadString('\n')
		cmd = strings.TrimSpace(cmd)
		if len(cmd) == 0 {
			continue
		}
		if strings.ToUpper(cmd) == "QUIT" {
			break
		}
		// 将获取到的数据发送至服务端
		if err := stream.Send(&pb.HelloRequest{Name: cmd}); err != nil {
			log.Fatalf("c.BidiHello stream.Send(%v) failed: %v", cmd, err)
		}
	}
	_ = stream.CloseSend()
	<-waitc
}

func main() {
	flag.Parse()
	creds, _ := credentials.NewClientTLSFromFile("../server.crt", "")
	// 连接到 server 端，此处禁用安全传输
	conn, err := grpc.Dial(*addr,
		grpc.WithTransportCredentials(creds),
		// 客户端注册拦截器
		grpc.WithUnaryInterceptor(interceptor.UnaryInterceptor),
		grpc.WithStreamInterceptor(interceptor.StreamInterceptor))

	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	defer func(conn *grpc.ClientConn) {
		err := conn.Close()
		if err != nil {
			log.Fatalf("conn close failed: %v", err)
		}
	}(conn)

	c := pb.NewGreeterClient(conn)

	execSayHello(c)
	//execLotsOfReplies(c)
	//execLotsOfGreetings(c)
	//execBidiHello(c)
}
