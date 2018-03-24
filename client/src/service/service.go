package service

import (
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/go-kit/kit/log"

	"pkg/addservice"
	"pkg/addtransport"
	"proto"
	//"github.com/hashicorp/consul/api"
)

type Pool struct{
	consulAddress string //如："127.0.0.1:8500"
	services map[string]addservice.Service
}

func NewPool(consulAddress string) *Pool {
	pool := &Pool{
		consulAddress:consulAddress,
		services:make(map[string]addservice.Service),
	}
	return pool
}

func (p *Pool) process() {

}

func (p *Pool) Get(serviceName string) addservice.Service {
	// 这里地址的来源要做成服务发现的形式
	// 通过服务名称去consul查找服务列表
	// 然后通过服务列表负载均衡的方式去分配
	// 这里是否需要经过网关 ？
	grpcAddr       := "0.0.0.0:8081"



	//分布式链路追踪
	zipkinV2URL    := "http://localhost:9411/api/v2/spans"
	zipkinV1URL    := "http://localhost:9411/api/v1/spans"


	// This is a demonstration client, which supports multiple tracers.
	// Your clients will probably just use one tracer.
	var otTracer stdopentracing.Tracer
	{
		collector, err := zipkinot.NewHTTPCollector(zipkinV1URL)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		defer collector.Close()
		var (
			debug       = false
			hostPort    = "localhost:0"
			serviceName = "addsvc-cli"
		)
		recorder := zipkinot.NewRecorder(collector, debug, hostPort, serviceName)
		otTracer, err = zipkinot.NewTracer(recorder)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}

	// This is a demonstration of the native Zipkin tracing client. If using
	// Zipkin this is the more idiomatic client over OpenTracing.
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = "" // if host:port is unknown we can keep this empty
			serviceName   = "addsvc-cli"
			useNoopTracer = false// (zipkinV2URL == "")
			reporter      = zipkinhttp.NewReporter(zipkinV2URL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			fmt.Fprintf(os.Stderr, "unable to create zipkin tracer: %s\n", err.Error())
			os.Exit(1)
		}
	}

	// This is a demonstration client, which supports multiple transports.
	// Your clients will probably just define and stick with 1 transport.
	var (
		svc addservice.Service
		err error
	)
	opt := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()))
	conn, err := grpc.Dial(grpcAddr, grpc.WithInsecure(), grpc.WithTimeout(time.Second), opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc = addtransport.NewGRPCClient(conn, otTracer, zipkinTracer, log.NewNopLogger())

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	return svc
}
