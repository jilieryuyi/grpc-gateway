package service

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"time"
	consul "github.com/hashicorp/consul/api"
	"sync"
)
const (
	Registered = 1 << iota
)
type Service struct {
	ServiceName string //service name, like: service.add
	ServiceHost string //service host, like: 0.0.0.0, 127.0.0.1
	ServiceIp string // if ServiceHost is 0.0.0.0, ServiceIp must set,
	// like 127.0.0.1 or 192.168.9.12 or 114.55.56.168
	ServicePort int // service port, like: 9998
	Interval time.Duration // interval for update ttl
	Ttl int //check ttl
	ServiceID string //serviceID = fmt.Sprintf("%s-%s-%d", name, ip, port)

	client *consul.Client ///consul client
	agent *consul.Agent //consul agent
	status int // register status
	lock *sync.Mutex //sync lock
}

type ServiceOption func(s *Service)

// set ttl
func Ttl(ttl int) ServiceOption {
	return func(s *Service){
		s.Ttl = ttl
	}
}

// set interval
func Interval(interval time.Duration) ServiceOption {
	return func(s *Service){
		s.Interval = interval
	}
}

// set service ip
func ServiceIp(serviceIp string) ServiceOption {
	return func(s *Service){
		s.ServiceIp = serviceIp
	}
}

// new a service
// name: service name
// host: service host like 0.0.0.0 or 127.0.0.1
// port: service port, like 9998
// consulAddress: consul service address, like 127.0.0.1:8500
// opts: ServiceOption, like ServiceIp("127.0.0.1")
// return new service pointer
func NewService(name string, host string, port int,
	consulAddress string, opts ...ServiceOption) *Service {
	sev := &Service{
		ServiceName:name,
		ServiceHost:host,
		ServicePort:port,
		Interval:time.Second * 10,
		Ttl:15,
		status:0,
		lock:new(sync.Mutex),
	}
	if len(opts) > 0 {
		for _, opt := range opts {
			opt(sev)
		}
	}
	conf := &consul.Config{Scheme: "http", Address: consulAddress}
	c, err := consul.NewClient(conf)
	if err != nil {
		log.Panicf("%v", err)
		return nil
	}
	sev.client = c
	ip := host
	if ip == "0.0.0.0" {
		if sev.ServiceIp == "" {
			log.Panicf("please set consul service ip")
		}
		ip = sev.ServiceIp
	}
	sev.ServiceID = fmt.Sprintf("%s-%s-%d", name, ip, port)
	sev.agent = sev.client.Agent()
	return sev
}

func (sev *Service) Deregister() error {
	err := sev.agent.ServiceDeregister(sev.ServiceID)
	if err != nil {
		log.Errorf("deregister service error: ", err.Error())
		return err
	}
	err = sev.agent.CheckDeregister(sev.ServiceID)
	if err != nil {
		log.Println("deregister check error: ", err.Error())
	}
	return err
}

func (sev *Service) Register() error {
	//de-register if meet signhup
	sev.lock.Lock()
	if sev.status & Registered <= 0 {
		sev.status |= Registered
	} else {
		sev.lock.Unlock()
		return nil
	}
	sev.lock.Unlock()
	// routine to update ttl
	go func() {
		ticker := time.NewTicker(sev.Interval)
		for {
			<-ticker.C
			err := sev.agent.UpdateTTL(sev.ServiceID, "", "passing")
			if err != nil {
				log.Println("update ttl of service error: ", err.Error())
			}
		}
	}()
	// initial register service
	ip := sev.ServiceHost
	if ip == "0.0.0.0" {
		ip = sev.ServiceIp
	}
	regis := &consul.AgentServiceRegistration{
		ID:      sev.ServiceID,
		Name:    sev.ServiceName,
		Address: ip,
		Port:    sev.ServicePort,
	}
	err := sev.agent.ServiceRegister(regis)
	if err != nil {
		return fmt.Errorf("initial register service '%s' host to consul error: %s", sev.ServiceName, err.Error())
	}
	// initial register service check
	check := consul.AgentServiceCheck{TTL: fmt.Sprintf("%ds", sev.Ttl), Status: "passing"}
	err = sev.agent.CheckRegister(&consul.AgentCheckRegistration{
		ID: sev.ServiceID,
		Name: sev.ServiceName,
		ServiceID: sev.ServiceID,
		AgentServiceCheck: check,
		})
	if err != nil {
		return fmt.Errorf("initial register service check to consul error: %s", err.Error())
	}
	return nil
}


func (sev *Service) Close() {
	sev.Deregister()
}