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
	"time"
	"google.golang.org/grpc"
	"github.com/jilieryuyi/grpc-gateway/proto"
	"github.com/jilieryuyi/grpc-gateway/service"
)

func main() {

	consulAddress := "127.0.0.1:8500"

	ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
	opt    := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()))
	r      := service.NewResolver(consulAddress)
	b      := grpc.RoundRobin(r)
	//wrapper
	//没有api可以初始化balancerWrapperBuilder，只有WithBalancer
	//虽然被Deprecated，但是也只能用WithBalancer了
	lb     := grpc.WithBalancer(b)

	conn, err := grpc.DialContext(ctx, "service.gateway", grpc.WithInsecure(), opt, lb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc := proto.NewServiceAddClient(conn)
	req := &proto.SumRequest{
		A:"100",
		B:"100",
	}
	v, err := svc.Sum(context.Background(), req, grpc.FailFast(false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%v + %v = %d  %v\n", req.A, req.B, v.V, v.Err)
}
