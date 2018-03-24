package main

import (
	"context"
	"fmt"
	"os"
	"service"
)


func main() {
	serviceName := "service.add"
	// service pool
	// 服务池
	pool := service.NewPool("127.0.0.1")
	svc  := pool.Get(serviceName)
	a := 100
	b := 100
	v, err := svc.Sum(context.Background(), int(a), int(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d + %d = %d\n", a, b, v)

}