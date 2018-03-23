package main

// grpc gateway 测试
// 运行server
// 运行 gateway
// 测试http://localhost:8083/sum?a=1&b=2000
// 这个是 http协议 转 grpc的代理
// grpc 透传 + http 转grpc实现

import (
	"net/http"
	"golang.org/x/net/context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
	"github.com/jilieryuyi/go-kit-grpc-test-demo/server/src/pb"
	"proxy"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"github.com/oklog/oklog/pkg/group"
	"net"
	"fmt"
	"os"
)

var (
	echoEndpoint = "localhost:8082"
)


type clientConn struct{
	client *grpc.ClientConn
}

func main() {
	//flag.Parse()
	//defer glog.Flush()

	var g group.Group
	{
		lis, _ := net.Listen("tcp", "0.0.0.0:8081")
		var connects = make(map[string]*clientConn)
		//proxy grpc server
		g.Add(func() error {
			//("proxy", "gRPC", "addr", "0.0.0.0:8081")
			// we add the Go Kit gRPC Interceptor to our gRPC service as it is used by
			// the here demonstrated zipkin tracing middleware.
			// todo 这里需要实现一套连接池
			//

			var director = func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {

				// Make sure we never forward internal services.
				//if strings.HasPrefix(fullMethodName, "/com.example.internal.") {
				//	return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
				//}
				md, ok := metadata.FromIncomingContext(ctx)
				// Copy the inbound metadata explicitly.
				outCtx, _ := context.WithCancel(ctx)
				outCtx = metadata.NewOutgoingContext(outCtx, md.Copy())
				if ok {
					// Decide on which backend to dial
					//if val, exists := md[":authority"]; exists && val[0] == "staging.api.example.com" {
						// Make sure we use DialContext so the dialing can be cancelled/time out together with the context.

					cl, ok := connects[echoEndpoint]
					if ok {
						fmt.Println("use pool")
						return outCtx, cl.client, nil
					}

						fmt.Println("new client")
						conn, err := grpc.DialContext(ctx, echoEndpoint, grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proxy.Codec())), grpc.WithInsecure())//grpc.WithCodec(Codec()))
					connects[echoEndpoint] = &clientConn{
						client:conn,
					}
						return outCtx, conn, err
					//} else if val, exists := md[":authority"]; exists && val[0] == "api.example.com" {
					//	conn, err := grpc.DialContext(ctx, "api-service.prod.svc.local", grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proxy.Codec())))//grpc.WithCodec(Codec()))
					//	return outCtx, conn, err
					//}
				}
				return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")
			}
			ser := grpc.NewServer(
				grpc.CustomCodec(proxy.Codec()),
				grpc.UnknownServiceHandler(proxy.TransparentHandler(director)))

			return ser.Serve(lis)
		}, func(error) {
			lis.Close()
		})
	}
	//{
	//	// This function just sits and waits for ctrl-C.
	//	cancelInterrupt := make(chan struct{})
	//	g.Add(func() error {
	//		c := make(chan os.Signal, 1)
	//		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
	//		select {
	//		case sig := <-c:
	//			return fmt.Errorf("received signal %s", sig)
	//		case <-cancelInterrupt:
	//			return nil
	//		}
	//	}, func(error) {
	//		close(cancelInterrupt)
	//	})
	//}

	{
		// http协议转grpc协议服务
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", ":8083")
		if err != nil {
			//logger.Log("transport", "debug/HTTP", "during", "Listen", "err", err)
			os.Exit(1)
		}

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := runtime.NewServeMux()
		opts := []grpc.DialOption{grpc.WithInsecure()}
		err = pb.RegisterAddHandlerFromEndpoint(ctx, mux, echoEndpoint, opts)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}

		//return http.ListenAndServe(":8083", mux)


		g.Add(func() error {
			//logger.Log("transport", "debug/HTTP", "addr", debugAddr)
			return http.Serve(debugListener, mux)
		}, func(error) {
			debugListener.Close()
		})
	}

	g.Run()

	//if err := run(); err != nil {
	//	glog.Fatal(err)
	//}
}
