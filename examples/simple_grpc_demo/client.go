package main

// run: go run client.go api.pb.go

import (
	"context"
	"fmt"
	"os"
	"time"
	"google.golang.org/grpc"
)

func main() {
	grpcAddr       := "127.0.0.1:8083"
	ctx, _ := context.WithTimeout(context.Background(), time.Second)
	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc := NewAddClient(conn)
	req := &SumRequest{
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
