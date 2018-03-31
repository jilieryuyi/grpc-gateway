package transport

import (
	"context"

	stdopentracing "github.com/opentracing/opentracing-go"
	stdzipkin "github.com/openzipkin/zipkin-go"
	oldcontext "golang.org/x/net/context"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/tracing/opentracing"
	"github.com/go-kit/kit/tracing/zipkin"
	grpctransport "github.com/go-kit/kit/transport/grpc"

	pb "github.com/jilieryuyi/grpc-gateway/proto"
	pendpoint "github.com/jilieryuyi/grpc-gateway/protocol/endpoint"
	"strconv"
	//"github.com/go-kit/kit/ratelimit"
	//"golang.org/x/time/rate"
	//"time"
	//"github.com/go-kit/kit/circuitbreaker"
	//"github.com/sony/gobreaker"
	"github.com/jilieryuyi/grpc-gateway/protocol/service"
	"google.golang.org/grpc"
	"github.com/go-kit/kit/endpoint"
	"fmt"
	"github.com/jilieryuyi/grpc-gateway/tools"
	"errors"
)

type grpcServer struct {
	sum    grpctransport.Handler
	concat grpctransport.Handler
}

// NewGRPCServer makes a set of endpoints available as a gRPC AddServer.
func NewGRPCServer(endpoints pendpoint.Set, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) pb.ServiceAddServer {
	// Zipkin GRPC Server Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing service can be instantiated
	// without an operation name and fed to each Go kit gRPC server as a
	// ServerOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path if used in combination with the Go kit gRPC Interceptor.
	//
	// In this example, we demonstrate a global Zipkin tracing service with
	// Go kit gRPC Interceptor.
	zipkinServer := zipkin.GRPCServerTrace(zipkinTracer)

	options := []grpctransport.ServerOption{
		grpctransport.ServerErrorLogger(logger),
		zipkinServer,
	}

	return &grpcServer{
		sum: grpctransport.NewServer(
			endpoints.SumEndpoint,
			decodeGRPCSumRequest,
			encodeGRPCSumResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "Sum", logger)))...,
		),
		concat: grpctransport.NewServer(
			endpoints.ConcatEndpoint,
			decodeGRPCConcatRequest,
			encodeGRPCConcatResponse,
			append(options, grpctransport.ServerBefore(opentracing.GRPCToContext(otTracer, "Concat", logger)))...,
		),
	}
}


// NewGRPCClient returns an AddService backed by a gRPC server at the other end
// of the conn. The caller is responsible for constructing the conn, and
// eventually closing the underlying transport. We bake-in certain middlewares,
// implementing the client library pattern.
func NewGRPCClient(conn *grpc.ClientConn, otTracer stdopentracing.Tracer, zipkinTracer *stdzipkin.Tracer, logger log.Logger) service.Service {
	// We construct a single ratelimiter middleware, to limit the total outgoing
	// QPS from this client to all methods on the remote instance. We also
	// construct per-endpoint circuitbreaker middlewares to demonstrate how
	// that's done, although they could easily be combined into a single breaker
	// for the entire remote instance, too.
	//limiter := ratelimit.NewErroringLimiter(rate.NewLimiter(rate.Every(time.Second), 100))

	// Zipkin GRPC Client Trace can either be instantiated per gRPC method with a
	// provided operation name or a global tracing client can be instantiated
	// without an operation name and fed to each Go kit client as ClientOption.
	// In the latter case, the operation name will be the endpoint's grpc method
	// path.
	//
	// In this example, we demonstrace a global tracing client.
	//zipkinClient := zipkin.GRPCClientTrace(zipkinTracer)

	// global client middlewares
	//options := []grpctransport.ClientOption{
	//	zipkinClient,
	//}

	// Each individual endpoint is an http/transport.Client (which implements
	// endpoint.Endpoint) that gets wrapped with various middlewares. If you
	// made your own client library, you'd do this work there, so your server
	// could rely on a consistent set of client behavior.
	var sumEndpoint endpoint.Endpoint
	{
		sumEndpoint = grpctransport.NewClient(
			conn,
			"proto.ServiceAdd",
			"Sum",
			encodeGRPCSumRequest,
			decodeGRPCSumResponse,
			pb.SumReply{},
			//append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		//sumEndpoint = opentracing.TraceClient(otTracer, "Sum")(sumEndpoint)
		//sumEndpoint = limiter(sumEndpoint)
		//sumEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
		//	Name:    "Sum",
		//	Timeout: 30 * time.Second,
		//}))(sumEndpoint)
	}

	// The Concat endpoint is the same thing, with slightly different
	// middlewares to demonstrate how to specialize per-endpoint.
	var concatEndpoint endpoint.Endpoint
	{
		concatEndpoint = grpctransport.NewClient(
			conn,
			"proto.ServiceAdd",
			"Concat",
			encodeGRPCConcatRequest,
			decodeGRPCConcatResponse,
			pb.ConcatReply{},
			//append(options, grpctransport.ClientBefore(opentracing.ContextToGRPC(otTracer, logger)))...,
		).Endpoint()
		//concatEndpoint = opentracing.TraceClient(otTracer, "Concat")(concatEndpoint)
		//concatEndpoint = limiter(concatEndpoint)
		//concatEndpoint = circuitbreaker.Gobreaker(gobreaker.NewCircuitBreaker(gobreaker.Settings{
		//	Name:    "Concat",
		//	Timeout: 10 * time.Second,
		//}))(concatEndpoint)
	}

	// Returning the endpoint.Set as a service.Service relies on the
	// endpoint.Set implementing the Service methods. That's just a simple bit
	// of glue code.
	return pendpoint.Set{
		SumEndpoint:    sumEndpoint,
		ConcatEndpoint: concatEndpoint,
	}
}

func (s *grpcServer) Sum(ctx oldcontext.Context, req *pb.SumRequest) (*pb.SumReply, error) {
	fmt.Printf("%v\n\n", ctx)
	h := tools.NewHeader(ctx)
	fmt.Printf("\n\nheader: %+v\n\n", *h)
	_, rep, err := s.sum.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.SumReply), nil
}

func (s *grpcServer) Concat(ctx oldcontext.Context, req *pb.ConcatRequest) (*pb.ConcatReply, error) {
	_, rep, err := s.concat.ServeGRPC(ctx, req)
	if err != nil {
		return nil, err
	}
	return rep.(*pb.ConcatReply), nil
}


// decodeGRPCSumRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC sum request to a user-domain sum request. Primarily useful in a server.
func decodeGRPCSumRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.SumRequest)
	a, _ := strconv.ParseInt(req.A, 10, 64)
	b, _ := strconv.ParseInt(req.B, 10, 64)
	return pendpoint.SumRequest{A: int(a), B: int(b)}, nil
}

// decodeGRPCConcatRequest is a transport/grpc.DecodeRequestFunc that converts a
// gRPC concat request to a user-domain concat request. Primarily useful in a
// server.
func decodeGRPCConcatRequest(_ context.Context, grpcReq interface{}) (interface{}, error) {
	req := grpcReq.(*pb.ConcatRequest)
	return pendpoint.ConcatRequest{A: req.A, B: req.B}, nil
}


// encodeGRPCSumResponse is a transport/grpc.EncodeResponseFunc that converts a
// user-domain sum response to a gRPC sum reply. Primarily useful in a server.
func encodeGRPCSumResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(pendpoint.SumResponse)
	return &pb.SumReply{V: int64(resp.V), Err: err2str(resp.Err)}, nil
}

// encodeGRPCConcatResponse is a transport/grpc.EncodeResponseFunc that converts
// a user-domain concat response to a gRPC concat reply. Primarily useful in a
// server.
func encodeGRPCConcatResponse(_ context.Context, response interface{}) (interface{}, error) {
	resp := response.(pendpoint.ConcatResponse)
	return &pb.ConcatReply{V: resp.V, Err: err2str(resp.Err)}, nil
}

// encodeGRPCSumRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain sum request to a gRPC sum request. Primarily useful in a client.
func encodeGRPCSumRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(pendpoint.SumRequest)
	//a ,_ := strconv.ParseInt(req.A, 10, 64)
	return &pb.SumRequest{A: fmt.Sprintf("%v",req.A), B: fmt.Sprintf("%v", req.B)}, nil
}

// encodeGRPCConcatRequest is a transport/grpc.EncodeRequestFunc that converts a
// user-domain concat request to a gRPC concat request. Primarily useful in a
// client.
func encodeGRPCConcatRequest(_ context.Context, request interface{}) (interface{}, error) {
	req := request.(pendpoint.ConcatRequest)
	return &pb.ConcatRequest{A: req.A, B: req.B}, nil
}

// decodeGRPCConcatResponse is a transport/grpc.DecodeResponseFunc that converts
// a gRPC concat reply to a user-domain concat response. Primarily useful in a
// client.
func decodeGRPCConcatResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.ConcatReply)
	return pendpoint.ConcatResponse{V: reply.V, Err: str2err(reply.Err)}, nil
}

// decodeGRPCSumResponse is a transport/grpc.DecodeResponseFunc that converts a
// gRPC sum reply to a user-domain sum response. Primarily useful in a client.
func decodeGRPCSumResponse(_ context.Context, grpcReply interface{}) (interface{}, error) {
	reply := grpcReply.(*pb.SumReply)
	return pendpoint.SumResponse{V: int(reply.V), Err: str2err(reply.Err)}, nil
}


func err2str(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func str2err(s string) error {
	if s == "" {
		return nil
	}
	return errors.New(s)
}