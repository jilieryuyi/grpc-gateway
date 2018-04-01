package main


// run: go run client.go
// 使用consul服务发现实现负载均衡
// 这里的"service.gateway"直接连接网关透传
// 也可以直连服务，比如将"service.gateway"修改为"service.add"即是直连服务
// 简单的grpc客户端实现

import (
	"context"
	"fmt"
	"os"
	"google.golang.org/grpc"
	"github.com/jilieryuyi/grpc-gateway/proto"
	"github.com/jilieryuyi/grpc-gateway/tools"
)

func main() {

	consulAddress := "127.0.0.1:8500"
	client := proto.NewClient(consulAddress)
	//客户端简单的调用流程即是proto.NewClient创建一个客户端
	//proto.NewServiceAddClient 来自于proto文件，由protoc自动生成
	//创建一个服务客户端
	svc    := proto.NewServiceAddClient(client.GetGrpcClient())

	req := &proto.SumRequest{
		A:"100",
		B:"100",
	}

	h := tools.NewHeader(context.Background())
	// client端发送header是通过context发送的
	ctx := h.Set("server", "grpc-client").ClientContext()
	v, err := svc.Sum(ctx, req, grpc.FailFast(false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%v + %v = %d  %v\n", req.A, req.B, v.V, v.Err)
}
