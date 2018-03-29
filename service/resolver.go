package service

import (
	log "github.com/sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
	//"google.golang.org/grpc"
	//"github.com/jilieryuyi/grpc-gateway/proto"
	"fmt"
)

// ConsulResolver is the implementaion of grpc.naming.Resolver
type ConsulResolver struct {
}

// NewResolver return ConsulResolver with service name
func NewResolver() *ConsulResolver {
	res := &ConsulResolver{}
	return res
}

// Resolve to resolve the service from consul, target is the dial address of consul
func (cr *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	fmt.Printf("Resolve was called======>%v\n", target)
	//return cr.watch, nil

	//r := service.NewResolver(grpcAddr, "127.0.0.1:8500")
	//b := grpc.RoundRobin(r)
	//
	//conn, err := grpc.Dial("service.add",//grpcAddr,
	//	grpc.WithInsecure(),
	//	grpc.WithTimeout(time.Second), opt,
	//	grpc.WithBalancer(b))

	//opt := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()))
	//cc, err := grpc.Dial(cr.grpcAddr, opt, grpc.WithInsecure())
	//if err != nil {
	//	return nil, err
	//}

	// generate consul client, return if error
	conf := &consul.Config{
		Scheme:  "http",
		Address: "127.0.0.1:8500",
	}
	client, err := consul.NewClient(conf)
	if err != nil {
		log.Panicf("creat consul error: %v", err)
	}
	// return ConsulWatcher
	//res.client = client

	return &ConsulWatcher{
		cr:cr,
		cc: client,
		target:target,
	}, nil

}
