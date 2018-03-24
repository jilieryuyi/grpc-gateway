package service

type Service struct {
	GrpcListenIp string//:= "0.0.0.0" //grpc监听ip
	GrpcListenPort int// := 8082 //grpc监听端口
	GrpcServiceIp string// := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
	ServiceName string
}

func NewService(GrpcListenIp string,
GrpcListenPort int,// := 8082 //grpc监听端口
GrpcServiceIp string,// := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
ServiceName string) *Service {
	sev := &Service{
		GrpcListenIp:GrpcListenIp,// string//:= "0.0.0.0" //grpc监听ip
		GrpcListenPort:GrpcListenPort,// string// := 8082 //grpc监听端口
		GrpcServiceIp:GrpcServiceIp,// string// := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
		ServiceName:ServiceName,// string
	}
	return sev
}

func (sev *Service) Register() {}
