package service

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"hello-server/pb"
)

func RegisterGreeterServer(s *grpc.Server) {
	pb.RegisterGreeterServer(s, &server{count: make(map[string]int)})
}

type server struct {
	pb.UnimplementedGreeterServer
	mu    sync.Mutex     // count 的并发锁
	count map[string]int // 记录每个 name 的请求次数
}

func (s *server) SayHello(ctx context.Context, req *pb.HelloRequest) (rsp *pb.HelloResponse, _ error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.count[req.Name]++ // 记录用户的请求次数
	// 超过1次就返回错误
	if s.count[req.Name] > 1 {
		st := status.New(codes.ResourceExhausted, "Request limit exceeded.")
		ds, err := st.WithDetails(
			&errdetails.QuotaFailure{
				Violations: []*errdetails.QuotaFailure_Violation{{
					Subject:     fmt.Sprintf("name:%s", req.Name),
					Description: "限制每个name调用一次",
				}},
			},
		)
		if err != nil {
			return nil, st.Err()
		}
		return nil, ds.Err()
	}
	// 正常返回响应
	return &pb.HelloResponse{Message: "Hello, " + req.Name}, nil
}

func (s *server) LotsOfReplies(req *pb.HelloRequest, stream pb.Greeter_LotsOfRepliesServer) error {
	words := []string{
		"你好",
		"hello",
		"こんにちは",
		"안녕하세요",
		"Здравствыйте",
	}
	for _, word := range words {
		data := &pb.HelloResponse{
			Message: word + req.GetName(),
		}
		// 使用 Send 方法返回多个数据
		if err := stream.Send(data); err != nil {
			return err
		}
	}
	return nil
}

func (s *server) LotsOfGreetings(stream pb.Greeter_LotsOfGreetingsServer) error {
	reply := "你好："
	for {
		// 接受客户端发来的流式数据
		req, err := stream.Recv()
		if err == io.EOF {
			// 最终统一回复
			return stream.SendAndClose(&pb.HelloResponse{
				Message: reply,
			})
		}
		if err != nil {
			return err
		}
		reply += req.GetName()
	}
}

// BidiHello 双向流式打招呼
func (s *server) BidiHello(stream pb.Greeter_BidiHelloServer) error {
	for {
		// 接受流式数据
		req, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		reply := magic(req.GetName()) // 对收到的数据做处理

		// 返回流式数据
		if err := stream.Send(&pb.HelloResponse{Message: reply}); err != nil {
			return err
		}
	}
}

// magic 一段价值连城的“人工智能”代码
func magic(s string) string {
	s = strings.ReplaceAll(s, "吗", "")
	s = strings.ReplaceAll(s, "吧", "")
	s = strings.ReplaceAll(s, "你", "我")
	s = strings.ReplaceAll(s, "？", "!")
	s = strings.ReplaceAll(s, "?", "!")
	return s
}
