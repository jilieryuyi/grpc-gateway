package service

import (
	log "github.com/sirupsen/logrus"
	"github.com/hashicorp/consul/api"
)

type Service struct {
	GrpcListenIp string
	GrpcListenPort int
	GrpcServiceIp string
	ServiceName string
	Client *api.Client
	Agent *api.Agent
}

var DefaultService = &Service{}

func NewService(GrpcListenIp string, GrpcListenPort int,
	GrpcServiceIp string, ServiceName string, ConsulAddress string) *Service {
	sev := &Service{
		GrpcListenIp:GrpcListenIp,
		GrpcListenPort:GrpcListenPort,
		GrpcServiceIp:GrpcServiceIp,
		ServiceName:ServiceName,
	}

	var err error
	consulConfig := api.DefaultConfig()
	consulConfig.Address = ConsulAddress
	sev.Client, err = api.NewClient(consulConfig)
	if err != nil {
		log.Panicf("create consul session with error: %+v", err)
	}
	sev.Agent = sev.Client.Agent()
	return sev
}

func (sev *Service) Register() {

}

func (sev *Service) GetServices() ([]*api.AgentService, error) {
	services, err := sev.Agent.Services()
	var sers []*api.AgentService
	for _, v := range services {
		if v.Service == sev.ServiceName {
			sers = append(sers, v)
		}
	}
	return sers, err
}