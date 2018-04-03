package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"github.com/oklog/oklog/pkg/group"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	addpb "github.com/jilieryuyi/grpc-gateway/proto"
	consul "github.com/jilieryuyi/grpc-gateway/service"
)

func main() {

	// 这里选项后续会放到配置文件
	grpcListenIp   := "0.0.0.0" //grpc监听ip
	grpcListenPort := 8182 //grpc监听端口
	grpcServiceIp  := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
	serviceName    := "service.add"
	consulAddress  := "127.0.0.1:8500"
	debugAddr      := "0.0.0.0:8080" // for prometheus
	zipkinAddress  := "localhost:9411"
	zipkinV2URL    := "http://" + zipkinAddress + "/api/v2/spans"
	zipkinV1URL    := "http://" + zipkinAddress + "/api/v1/spans"

	sev := consul.NewService(serviceName, grpcListenIp, grpcListenPort, consulAddress, consul.ServiceIp(grpcServiceIp))
	sev.Register()
	defer sev.Close()

	helper := NewHelper(grpcListenIp,grpcListenPort,
			grpcServiceIp, serviceName, consulAddress,
			debugAddr, zipkinAddress, zipkinV2URL, zipkinV1URL)
	defer helper.Close()
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	var g group.Group
	{
		debugListener, err := net.Listen("tcp", debugAddr)
		if err != nil {
			helper.logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			helper.logger.Log("transport", "debug/HTTP", "addr", debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}

	{
		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort))
		if err != nil {
			helper.logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			helper.logger.Log("transport", "gRPC", "addr", fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort))

			baseServer := grpc.NewServer(
				grpc.UnaryInterceptor(kitgrpc.Interceptor),
				grpc.CustomCodec(addpb.Codec()))
			grpcServer := helper.getServer()
			addpb.RegisterServiceAddServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}

	{
		cancelInterrupt := make(chan struct{})
		g.Add(func() error {
			c := make(chan os.Signal, 1)
			signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
			select {
			case sig := <-c:
				return fmt.Errorf("received signal %s", sig)
			case <-cancelInterrupt:
				return nil
			}
		}, func(error) {
			close(cancelInterrupt)
		})
	}
	helper.logger.Log("exit", g.Run())
}