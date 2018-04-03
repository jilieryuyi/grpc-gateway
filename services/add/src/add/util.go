package main

import (
	"fmt"
	"os"

	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"

	addpb "github.com/jilieryuyi/grpc-gateway/proto"
	"github.com/jilieryuyi/grpc-gateway/protocol/endpoint"
	"github.com/jilieryuyi/grpc-gateway/protocol/service"
	"github.com/jilieryuyi/grpc-gateway/protocol/transport"
	"github.com/openzipkin/zipkin-go/reporter"
)

type Helper struct {
	grpcListenIp string//   := "0.0.0.0" //grpc监听ip
	grpcListenPort int// := 8182 //grpc监听端口
	grpcServiceIp string//  := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
	serviceName string//    := "service.add"
	consulAddress string//  := "127.0.0.1:8500"
	debugAddr string//      := "0.0.0.0:8080" // for prometheus
	zipkinAddress string//  := "localhost:9411"
	zipkinV2URL string//    := "http://" + zipkinAddress + "/api/v2/spans"
	zipkinV1URL string//    := "http://" + zipkinAddress + "/api/v1/spans"
	logger log.Logger
	collector zipkinot.Collector
	reporter reporter.Reporter
	ints metrics.Counter
	chars metrics.Counter
	duration metrics.Histogram
}

func NewHelper(grpcListenIp string,//   := "0.0.0.0" //grpc监听ip
	grpcListenPort int,// := 8182 //grpc监听端口
	grpcServiceIp string,//  := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
	serviceName string,//    := "service.add"
	consulAddress string,//  := "127.0.0.1:8500"
	debugAddr string,//      := "0.0.0.0:8080" // for prometheus
	zipkinAddress string,//  := "localhost:9411"
	zipkinV2URL string,//    := "http://" + zipkinAddress + "/api/v2/spans"
	zipkinV1URL string) *Helper {

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	h := &Helper{
		grpcListenIp:grpcListenIp,// string//   := "0.0.0.0" //grpc监听ip
		grpcListenPort:grpcListenPort,//, int// := 8182 //grpc监听端口
		grpcServiceIp:grpcServiceIp,// string//  := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
		serviceName:serviceName,// string//    := "service.add"
		consulAddress:consulAddress,// string//  := "127.0.0.1:8500"
		debugAddr:debugAddr,// string//      := "0.0.0.0:8080" // for prometheus
		zipkinAddress:zipkinAddress,// string//  := "localhost:9411"
		zipkinV2URL:zipkinV2URL,// string//    := "http://" + zipkinAddress + "/api/v2/spans"
		zipkinV1URL:zipkinV1URL,// string//
		logger:logger,
	}
	var err error
	h.collector, err = zipkinot.NewHTTPCollector(h.zipkinV1URL)
	if err != nil {
		h.logger.Log("err", err)
		os.Exit(1)
	}
	h.reporter      = zipkinhttp.NewReporter(zipkinV2URL)
	h.ints, h.chars = h.metrics()
	h.duration = h.getDuration()
	return h
}

func (h *Helper) Close() {
	h.collector.Close()
	h.reporter.Close()
}

func (h *Helper) tracer() stdopentracing.Tracer {
	var tracer stdopentracing.Tracer
	{
		var err error
		h.logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", h.zipkinV1URL)
		var (
			debug       = false
			hostPort    = fmt.Sprintf("%v:%v", h.grpcListenIp, h.grpcListenPort)
			//serviceName = "addsvc"
		)
		recorder := zipkinot.NewRecorder(h.collector, debug, hostPort, h.serviceName)
		tracer, err = zipkinot.NewTracer(recorder)
		if err != nil {
			h.logger.Log("err", err)
			os.Exit(1)
		}
	}
	return tracer
}

func (h *Helper) zipkinTracer() *zipkin.Tracer {
	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = fmt.Sprintf("%v:%v", h.grpcListenIp, h.grpcListenPort)//"localhost:80"
			//serviceName   = "addsvc"
			useNoopTracer = false//(zipkinV2URL == "")
		)
		zEP, _ := zipkin.NewEndpoint(h.serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			h.reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			h.logger.Log("err", err)
			os.Exit(1)
		}
	}
	return zipkinTracer
}

func (h *Helper) metrics()(ints, chars metrics.Counter) {
	//var ints, chars metrics.Counter
	{
		// Business-level metrics.
		ints = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "integers_summed",
			Help:      "Total count of integers summed via the Sum method.",
		}, []string{})
		chars = prometheus.NewCounterFrom(stdprometheus.CounterOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "characters_concatenated",
			Help:      "Total count of characters concatenated via the Concat method.",
		}, []string{})
	}
	return// ints, chars
}

func (h *Helper) getDuration() (duration metrics.Histogram) {
	//var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	return
}

func (h *Helper) getServer() addpb.ServiceAddServer {
	tracer := h.tracer()
	zipkinTracer := h.zipkinTracer()
	var (
		addService     = service.New(h.logger, h.ints, h.chars)
		endpoints      = endpoint.New(addService, h.logger, h.duration, tracer, zipkinTracer)
		grpcServer     = transport.NewGRPCServer(endpoints, tracer, zipkinTracer, h.logger)
	)
	return grpcServer
}
