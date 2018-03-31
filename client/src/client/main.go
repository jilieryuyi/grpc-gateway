package main

import (
	"context"
	"fmt"
	"os"
)

// go-kit客户端实现
// 使用consul服务发现，支持负载均衡

func main() {
	//分布式链路追踪
	zipkinV2URL    := "http://localhost:9411/api/v2/spans"
	zipkinV1URL    := "http://localhost:9411/api/v1/spans"
	consulAddress  := "127.0.0.1:8500"

	pool := NewPool(zipkinV2URL, zipkinV1URL, consulAddress)
	defer pool.Close()
	svc  := pool.getService()
	fmt.Printf("=====\n")
	a    := 100
	b    := 100
	v, err := svc.Sum(context.Background(), int(a), int(b))
	fmt.Printf("=====\n")

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "%d + %d = %d\n", a, b, v)

}