package main

// run: go run client.go api.pb.go
// 简单的grpc客户端实现

import (
	"context"
	"fmt"
	"os"
	"time"
	"google.golang.org/grpc"
	"github.com/jilieryuyi/go-kit-grpc-test-demo/server/src/pb"

)

func main() {
	grpcAddr       := "127.0.0.1:8081"
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc := pb.NewAddClient(conn)
	req := &pb.SumRequest{
		A:100,
		B:100,
	}
	v, err := svc.Sum(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d + %d = %d  %v\n", req.A, req.B, v.V, v.Err)
}
