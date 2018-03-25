package service

import (
	"errors"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
	log "github.com/sirupsen/logrus"
)
type resolver struct {
	serviceName string // service name to resolve
	 client *api.Client
}

// NewResolver return resolver with service name
func NewResolver(serviceName string, consulAddress string) *resolver {
	config := api.DefaultConfig()
	config.Address = consulAddress
	// generate etcd client
	client, err := api.NewClient(config)
	if err != nil {
		log.Panicf("%+v", err)
		//return nil, fmt.Errorf("grpclb: creat etcd3 client failed: %s", err.Error())
	}
	return &resolver{serviceName: serviceName, client:client}
}

func (re *resolver) Resolve(target string) (naming.Watcher, error) {
	if re.serviceName == "" {
		return nil, errors.New("grpclb: no service name provided")
	}
	// Return watcher
	return re, nil//&watcher{re: re, client: *client}, nil
}
