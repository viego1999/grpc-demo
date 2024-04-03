## gRPC-demo

基于 gRPC 实现的 demo 项目，快速搭建基于 gRPC 的微服务项目

主要技术：Go + protobuf + gRPC + gRPC-gateway + oauth2 + SSL/TLS

目录结构说明：

```
- grpc-demo
  - demo                // protobuf 的使用 demo 项目
  - hello-client        // grpc 客户端模块
    - interceptor       // 客户端过滤器
    - pb                // proto 文件
    - go.mod
    - main.go
  - hello-client-py     // grpc 客户端模块（python实现）
    - pb                // proto 文件
    - hello_pb2.py
    - hello_pb2_grpc.py
    - main.py
  - hello-server        // grpc 服务端模块
    - interceptor       // 服务端过滤器
    - pb                // proto 文件
    - service           // 服务端服务
    - go.mod
    - main.go
```

详细教程链接：https://www.liwenzhou.com/posts/Go/gRPC/