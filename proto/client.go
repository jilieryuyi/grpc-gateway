package proto
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
	"github.com/jilieryuyi/grpc-gateway/service"
)

type Client struct {
	consulAddress string
	client *grpc.ClientConn
}

//consulAddress like: 127.0.0.1:8500
func NewClient(consulAddress string) *Client{
	c := &Client{
		consulAddress:consulAddress,
	}
	c.init()
	return c
}

func (c *Client) init() {
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
	opt    := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(Codec()), grpc.FailFast(false))
	r      := service.NewResolver(c.consulAddress)
	b      := grpc.RoundRobin(r)
	//wrapper
	//没有api可以初始化balancerWrapperBuilder，只有WithBalancer
	//虽然被Deprecated，但是也只能用WithBalancer了
	lb     := grpc.WithBalancer(b)

	var err error
	c.client, err = grpc.DialContext(ctx, "service.gateway", grpc.WithInsecure(), opt, lb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
}

func (c *Client) GetGrpcClient() *grpc.ClientConn {
	return c.client
}

func (c *Client) Close() {
	c.client.Close()
}
