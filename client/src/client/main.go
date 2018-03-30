package main

import (
	"context"
	"fmt"
	"os"
	_ "time"

	"google.golang.org/grpc"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"

	"github.com/go-kit/kit/log"
	"github.com/jilieryuyi/grpc-gateway/service"
	"pkg/addservice"
	"pkg/addtransport"
	"github.com/jilieryuyi/grpc-gateway/proto"
)

// go-kit客户端实现
// 使用consul服务发现，支持负载均衡

func main() {
	//serviceName := "service.add"
	//// service pool
	//// 服务池
	//pool := service.NewPool("127.0.0.1:8500")
	//svc  := pool.Get(serviceName)
	// 这里地址的来源要做成服务发现的形式
	// 通过服务名称去consul查找服务列表
	// 然后通过服务列表负载均衡的方式去分配
	// 这里是否需要经过网关 ？
	//grpcAddr       := "0.0.0.0:8082"



	//分布式链路追踪
	zipkinV2URL    := "http://localhost:9411/api/v2/spans"
	zipkinV1URL    := "http://localhost:9411/api/v1/spans"
	consulAddress  := "127.0.0.1:8500"

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

	fmt.Printf("stdopentracing init\n")

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

	fmt.Printf("zipkin init\n")

	// This is a demonstration client, which supports multiple transports.
	// Your clients will probably just define and stick with 1 transport.
	var (
		svc addservice.Service
		err error
	)

	//以下部分实现了grpc负载均衡
	resolver := service.NewResolver(consulAddress)
	robin    := grpc.RoundRobin(resolver)
	lb       := grpc.WithBalancer(robin)
	//这个选项用于等待consul完成服务发现初始化
	cp       := grpc.WithDefaultCallOptions(grpc.FailFast(false))
	//这个选项用于设置grpc的编码解码实现
	opt      := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()))

	conn, err := grpc.Dial("service.gateway", grpc.WithInsecure(), opt, cp, lb)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	fmt.Printf("dial grpc\n")

	defer conn.Close()
	svc = addtransport.NewGRPCClient(conn, otTracer, zipkinTracer, log.NewNopLogger())
	fmt.Printf("new client\n")

	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("new client return\n")

	a := 100
	b := 100
	v, err := svc.Sum(context.Background(), int(a), int(b))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d + %d = %d\n", a, b, v)

}