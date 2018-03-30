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
	"google.golang.org/grpc"
	"proxy"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"github.com/oklog/oklog/pkg/group"
	"net"
	"fmt"
	"os"
	"github.com/jilieryuyi/grpc-gateway/service"
	"strings"
	consul "github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

func main() {

	// grpc gateway 代理服务
	grpcListenIp := "0.0.0.0"
	grpcListenPort := 8081
	consulAddress := "127.0.0.1:8500"
	grpcServiceIp := "127.0.0.1"

	//todo 注册gateway服务
	serviceName := "service.gateway"
	sev := service.NewService(serviceName, grpcListenIp, grpcListenPort, consulAddress, service.ServiceIp(grpcServiceIp))
	sev.Register()
	defer sev.Close()

	grpcAddr := fmt.Sprintf("%v:%v", grpcListenIp, grpcListenPort)//":8081"
	var g group.Group
	{
		lis, _ := net.Listen("tcp", grpcAddr)
		//var connects = make(map[string]*clientConn)
		var conns = make(map[string]*grpc.ClientConn)
		conf := consul.DefaultConfig()
		conf.Address = consulAddress//"127.0.0.1:8500"
		consulClient, _ := consul.NewClient(conf)


		//proxy grpc server
		g.Add(func() error {
			var director = func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {
				// /proto.ServiceAdd/Sum
				// todo 负载均衡

				ms := strings.Split(fullMethodName, "/")
				for i, v := range ms {
					fmt.Printf("%+v => %+v\n", i, v)
				}
				in := strings.Index(ms[1], ".")
				serviceName := strings.ToLower(ms[1][in+1:])
				fmt.Printf("serviceName=>%+v\n", serviceName)


				//var adds = make(map[string][]string)

				//Health().Service(cw.target, "", true, q)
				//for key, kc := range cs {
				//	fmt.Printf("%v==>%+v", key, *kc)
				//	sn := strings.Replace(kc.Service, ".", "", -1)
				//	//if serviceName == strings.ToLower(sn) {
				//		adds[strings.ToLower(sn)] = append(adds[strings.ToLower(sn)], fmt.Sprintf("%v:%v", kc.Address, kc.Port))
				//	//}
				//}
				//fmt.Printf("#####%+v, %+v, %+v",cs, meta, err)
				//if err != nil {
				//	return nil, 0, err
				//}

				//addrs := make([]string, 0)
				//for _, s := range cs {
				//	// addr should like: 127.0.0.1:8001
				//	addrs = append(addrs, fmt.Sprintf("%s:%d", s.Service.Address, s.Service.Port))
				//}
				//
				//return addrs, meta.LastIndex, nil





				fmt.Printf("%+v\n", fullMethodName)
				fmt.Printf("%+v\n", ctx)
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

					//cl, ok := connects[echoEndpoint]
					//if ok {
					//	fmt.Println("use pool")
					//	return outCtx, cl.client, nil
					//}
					//
					//fmt.Println("new client")

					//addresses := adds[serviceName]
					//r := rand.New(rand.NewSource(time.Now().UnixNano()))
					//msn := r.Intn(1000000)
					//index := msn%len(addresses)
					//ep := addresses[index]



					conn, ok := conns[serviceName]
					if ok && conn != nil {
						fmt.Printf("use pool: %v\n", serviceName)
						return outCtx, conn, nil
					}

					resl   := service.NewResolver(consulAddress)
					rr     := grpc.RoundRobin(resl)
					lb     := grpc.WithBalancer(rr)
					cs, _  := consulClient.Agent().Services()
					var err error
					for _, kc := range cs {
						if kc.Service == "service.gateway" {
							continue
						}
						sn := strings.Replace(kc.Service, ".", "", -1)
						n, ok := conns[strings.ToLower(sn)]
						if n != nil && ok {
							continue
						}
						//if serviceName == strings.ToLower(sn) {
						conns[strings.ToLower(sn)], err = grpc.DialContext(ctx, kc.Service, grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proxy.Codec()), grpc.FailFast(false)), grpc.WithInsecure(), lb) //grpc.WithCodec(Codec()))
						if err != nil {
							log.Errorf("%+v", err)
							conns[strings.ToLower(sn)] = nil
						}
					}

					conn, ok = conns[serviceName]
					if ok && conn != nil {
						return outCtx, conn, nil
					}

					//connects[echoEndpoint] = &clientConn{
					//	client:conn,
					//}
					//return nil, nil, status.Errorf(codes.NotFound, "service not found: " + serviceName)
					return nil, nil, status.Errorf(codes.Unimplemented, "Unknown method")

					//return outCtx, conns[serviceName], err
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

	{
		// http协议转grpc协议服务
		// The debug listener mounts the http.DefaultServeMux, and serves up
		// stuff like the Prometheus metrics route, the Go debug and profiling
		// routes, and so on.
		debugListener, err := net.Listen("tcp", ":8084")
		if err != nil {
			os.Exit(1)
		}

		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		mux := proxy.NewMyMux(ctx, consulAddress)//MyMux{}//runtime.NewServeMux()
		defer mux.Close()
		//opts := []grpc.DialOption{grpc.WithInsecure()}
		//err = proxy.RegisterHandlerFromEndpoint(ctx, mux, opts)
		if err != nil {
			fmt.Printf("%v\n", err)
			return
		}
		g.Add(func() error {
			return http.Serve(debugListener, mux)
		}, func(error) {
			debugListener.Close()
		})
	}
	g.Run()
}