package main

import (
	"fmt"
	"os"
	_ "time"
	"context"
	"google.golang.org/grpc"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/go-kit/kit/log"
	consul "github.com/jilieryuyi/grpc-gateway/service"
	"github.com/jilieryuyi/grpc-gateway/protocol/transport"
	"github.com/jilieryuyi/grpc-gateway/proto"
	"github.com/jilieryuyi/grpc-gateway/protocol/service"

	"time"
	"github.com/openzipkin/zipkin-go/reporter"
)

// go-kit客户端实现
// 使用consul服务发现，支持负载均衡

//zipkinV2URL    := "http://localhost:9411/api/v2/spans"
//zipkinV1URL    := "http://localhost:9411/api/v1/spans"

type Pool struct {
	zipkinV2URL string//    := "http://localhost:9411/api/v2/spans"
	zipkinV1URL string//    := "http://localhost:9411/api/v1/spans"
	consulAddress string//  := "127.0.0.1:8500"
	reporter reporter.Reporter
	collector zipkinot.Collector
}

func NewPool(zipkinV2URL string, zipkinV1URL string, consulAddress string) *Pool {
	p := &Pool{
		zipkinV2URL:zipkinV2URL,
		zipkinV1URL:zipkinV1URL,
		consulAddress:consulAddress,
	}
	return p
}

func (p *Pool) Close() {
	p.reporter.Close()
	p.collector.Close()
}

func (p *Pool) getOtTracer() stdopentracing.Tracer {
	var err error
	var otTracer stdopentracing.Tracer
	{
		p.collector, err = zipkinot.NewHTTPCollector(p.zipkinV1URL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		var (
			debug       = false
			hostPort    = "localhost:0"
			serviceName = "addsvc-cli"
		)
		recorder := zipkinot.NewRecorder(p.collector, debug, hostPort, serviceName)
		otTracer, err = zipkinot.NewTracer(recorder)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
	return otTracer
}

func (p *Pool) getZipkinTracer() *zipkin.Tracer {
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "" // if host:port is unknown we can keep this empty
			serviceName   = "addsvc-cli"
			useNoopTracer = false// (zipkinV2URL == "")
		)
		p.reporter      = zipkinhttp.NewReporter(p.zipkinV2URL)
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			p.reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to create zipkin tracer: %s\n", err.Error())
			os.Exit(1)
		}
	}
	return zipkinTracer
}

func (p *Pool) getGrpcClient() *grpc.ClientConn {
	ctx, _ := context.WithTimeout(context.Background(), time.Second * 3)
	opt    := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()), grpc.FailFast(false))
	r      := consul.NewResolver(p.consulAddress)
	b      := grpc.RoundRobin(r)
	//wrapper
	//没有api可以初始化balancerWrapperBuilder，只有WithBalancer
	//虽然被Deprecated，但是也只能用WithBalancer了
	lb     := grpc.WithBalancer(b)
	conn, err := grpc.DialContext(ctx, "service.gateway", grpc.WithInsecure(), opt, lb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	return conn
}

func (p *Pool) getService() service.Service {
	svc := transport.NewGRPCClient(p.getGrpcClient(), p.getOtTracer(), p.getZipkinTracer(), log.NewNopLogger())
	return svc
}