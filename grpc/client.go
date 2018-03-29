package main

// run: go run client.go api.pb.go
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
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
	opt    := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()))
	r      := service.NewResolver()
	b      := grpc.RoundRobin(r)
	lb     := grpc.WithBalancer(b)

	conn, err := grpc.DialContext(ctx, "service.add", grpc.WithInsecure(), opt, lb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc := proto.NewServiceAddClient(conn)
	req := &proto.SumRequest{
		A:100,
		B:100,
	}
	v, err := svc.Sum(context.Background(), req, grpc.FailFast(false))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d + %d = %d  %v\n", req.A, req.B, v.V, v.Err)
}
