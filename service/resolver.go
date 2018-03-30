package service

import (
	log "github.com/sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

// ConsulResolver is the implementaion of grpc.naming.Resolver
type ConsulResolver struct {
	client *consul.Client
}

// NewResolver return ConsulResolver with service name
func NewResolver(consulAddress string) *ConsulResolver {
	conf := &consul.Config{
		Scheme:  "http",
		Address: consulAddress,//"127.0.0.1:8500",
	}
	client, err := consul.NewClient(conf)
	if err != nil {
		log.Panicf("creat consul error: %v", err)
	}
	res := &ConsulResolver{
		client:client,
	}
	return res
}

// target is the service name like: service.gateway
func (cr *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return &ConsulWatcher{
		cr: cr,
		cc: cr.client,
		target: target,
		health:cr.client.Health(),
	}, nil
}
