package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/oklog/oklog/pkg/group"
	stdopentracing "github.com/opentracing/opentracing-go"
	zipkin "github.com/openzipkin/zipkin-go"
	zipkinot "github.com/openzipkin/zipkin-go-opentracing"
	zipkinhttp "github.com/openzipkin/zipkin-go/reporter/http"
	stdprometheus "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/prometheus"
	kitgrpc "github.com/go-kit/kit/transport/grpc"

	addpb "github.com/jilieryuyi/grpc-gateway/proto"
	"pkg/addendpoint"
	"pkg/addservice"
	"pkg/addtransport"

	"google.golang.org/grpc/encoding"
	eproto "google.golang.org/grpc/encoding/proto"

	"encoding/json"
	"github.com/jilieryuyi/grpc-gateway/service"
)



func MyCodec() grpc.Codec {
	defauleCode := encoding.GetCodec(eproto.Name)
	return &myprotoCodec{defauleCode, 0}
}

type myprotoCodec struct{
	DefauleCode encoding.Codec
	DataType int //0 default 1 json
}


// CodecWithParent returns a proxying grpc.Codec with a user provided codec as parent.
//
// This codec is *crucial* to the functioning of the proxy. It allows the proxy server to be oblivious
// to the schema of the forwarded messages. It basically treats a gRPC message frame as raw bytes.
// However, if the server handler, or the client caller are not proxy-internal functions it will fall back
// to trying to decode the message using a fallback codec.
//func MyCodecWithParent(fallback grpc.Codec) grpc.Codec {
//	return &rawCodec{fallback}
//}


func (d *myprotoCodec) Marshal(v interface{}) ([]byte, error) {
	fmt.Printf("myprotoCodec Marshal msg: %v\n", v)
	//if d.DataType == 1 {
	//	d.DataType = 0
		return json.Marshal(v)
	//}
	//return d.DefauleCode.Marshal(v)
}

func (d *myprotoCodec) Unmarshal(data []byte, v interface{}) error {
	fmt.Printf("myprotoCodec Unmarshal msg: %+v, %+v\n", string(data), v)
	//err := json.Unmarshal(data, &v)
	//if err != nil {
	//	return proto.Unmarshal(data, v.(proto.Message))
	//}
	//return err
	//prev := []byte("----json----")
	//if bytes.HasPrefix(data, prev) {
	//	d.DataType = 1
		return json.Unmarshal(data, v)
	//}
	//return d.DefauleCode.Unmarshal(data, v)
}

func (d *myprotoCodec) String() string {
	fmt.Println("myprotoCodec string")
	//return "proto"
	return  d.DefauleCode.Name()
}

func (d *myprotoCodec) Name() string {
	fmt.Println("myprotoCodec string")
	return  d.DefauleCode.Name()
}



func main() {

	// 这里选项后续会放到配置文件
	grpcListenIp := "0.0.0.0" //grpc监听ip
	grpcListenPort := 8082 //grpc监听端口
	grpcServiceIp := "127.0.0.1" //grpc通过这个ip对外服务，可以自由配置外网内网，配合监听ip使用
	serviceName := "service.add"
	consulAddress := "127.0.0.1:8500"
	debugAddr      := "0.0.0.0:8080" // for prometheus


	//这是一个服务
	//将被注册到consul
	sev := service.NewService(grpcListenIp, grpcListenPort,
		grpcServiceIp, serviceName, consulAddress)
	sev.Register()

	//grpc服务接听
	//分布式链路追踪
	zipkinV2URL    := "http://localhost:9411/api/v2/spans"
	zipkinV1URL    := "http://localhost:9411/api/v1/spans"

	// Create a single logger, which we'll use and give to other components.
	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
		logger = log.With(logger, "caller", log.DefaultCaller)
	}

	// Determine which OpenTracing tracer to use. We'll pass the tracer to all the
	// components that use it, as a dependency.
	var tracer stdopentracing.Tracer
	{
			logger.Log("tracer", "Zipkin", "type", "OpenTracing", "URL", zipkinV1URL)
			collector, err := zipkinot.NewHTTPCollector(zipkinV1URL)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
			defer collector.Close()
			var (
				debug       = false
				hostPort    = fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort)
				serviceName = "addsvc"
			)
			recorder := zipkinot.NewRecorder(collector, debug, hostPort, serviceName)
			tracer, err = zipkinot.NewTracer(recorder)
			if err != nil {
				logger.Log("err", err)
				os.Exit(1)
			}
	}

	var zipkinTracer *zipkin.Tracer
	{
		var (
			err           error
			hostPort      = fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort)//"localhost:80"
			serviceName   = "addsvc"
			useNoopTracer = false//(zipkinV2URL == "")
			reporter      = zipkinhttp.NewReporter(zipkinV2URL)
		)
		defer reporter.Close()
		zEP, _ := zipkin.NewEndpoint(serviceName, hostPort)
		zipkinTracer, err = zipkin.NewTracer(
			reporter, zipkin.WithLocalEndpoint(zEP), zipkin.WithNoopTracer(useNoopTracer),
		)
		if err != nil {
			logger.Log("err", err)
			os.Exit(1)
		}
	}

	// Create the (sparse) metrics we'll use in the service. They, too, are
	// dependencies that we pass to components that use them.
	var ints, chars metrics.Counter
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
	var duration metrics.Histogram
	{
		// Endpoint-level metrics.
		duration = prometheus.NewSummaryFrom(stdprometheus.SummaryOpts{
			Namespace: "example",
			Subsystem: "addsvc",
			Name:      "request_duration_seconds",
			Help:      "Request duration in seconds.",
		}, []string{"method", "success"})
	}
	http.DefaultServeMux.Handle("/metrics", promhttp.Handler())

	// Build the layers of the service "onion" from the inside out. First, the
	// business logic service; then, the set of endpoints that wrap the service;
	// and finally, a series of concrete transport adapters. The adapters, like
	// the HTTP handler or the gRPC server, are the bridge between Go kit and
	// the interfaces that the transports expect. Note that we're not binding
	// them to ports or anything yet; we'll do that next.
	var (
		service        = addservice.New(logger, ints, chars)
		endpoints      = addendpoint.New(service, logger, duration, tracer, zipkinTracer)
		grpcServer     = addtransport.NewGRPCServer(endpoints, tracer, zipkinTracer, logger)
	)

	// Now we're to the part of the func main where we want to start actually
	// running things, like servers bound to listeners to receive connections.
	//
	// The method is the same for each component: add a new actor to the group
	// struct, which is a combination of 2 anonymous functions: the first
	// function actually runs the component, and the second function should
	// interrupt the first function and cause it to return. It's in these
	// functions that we actually bind the Go kit server/handler structs to the
	// concrete transports and run them.
	//
	// Putting each component into its own block is mostly for aesthetics: it
	// clearly demarcates the scope in which each listener/socket may be used.
	var g group.Group
	{
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", debugAddr)
		if err != nil {
			logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "debug/HTTP", "addr", debugAddr)
			return http.Serve(debugListener, http.DefaultServeMux)
		}, func(error) {
			debugListener.Close()
		})
	}
	{
		// The gRPC listener mounts the Go kit gRPC server we created.
		grpcListener, err := net.Listen("tcp", fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort))
		if err != nil {
			logger.Log("transport", "gRPC", "during", "Listen", "err", err)
			os.Exit(1)
		}
		g.Add(func() error {
			logger.Log("transport", "gRPC", "addr", fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort))
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.


			baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor),
				grpc.CustomCodec(MyCodec()))

			addpb.RegisterAddServer(baseServer, grpcServer)
			return baseServer.Serve(grpcListener)
		}, func(error) {
			grpcListener.Close()
		})
	}

	{
		// This function just sits and waits for ctrl-C.
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
	logger.Log("exit", g.Run())
}