package service

import (
	log "github.com/sirupsen/logrus"
	consul "github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

// ConsulResolver is the implementaion of grpc.naming.Resolver
type ConsulResolver struct {
	ServiceName string //service name
	watch *ConsulWatcher
}

// NewResolver return ConsulResolver with service name
func NewResolver(serviceName string, consulAddress string) *ConsulResolver {
	if serviceName == "" {
		log.Panicf("no service name provided")
	}
	res := &ConsulResolver{ServiceName: serviceName}
	// generate consul client, return if error
	conf := &consul.Config{
		Scheme:  "http",
		Address: consulAddress,
	}
	client, err := consul.NewClient(conf)
	if err != nil {
		log.Panicf("creat consul error: %v", err)
	}
	// return ConsulWatcher
	res.watch = &ConsulWatcher{
		cr: res,
		cc: client,
	}
	return res
}

// Resolve to resolve the service from consul, target is the dial address of consul
func (cr *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return cr.watch, nil
}
