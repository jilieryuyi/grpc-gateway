package proxy

import (
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"fmt"
)

type clientConn struct{
	client *grpc.ClientConn
}

//连接池实现
type Pool struct{
	connects map[string]*clientConn
}

func (p *Pool) NewPool() {

}

func (p *Pool) StreamDirector() StreamDirector {
	return func(ctx context.Context, fullMethodName string) (context.Context, *grpc.ClientConn, error) {

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
			echoEndpoint := "localhost:8082"

			cl, ok := p.connects[echoEndpoint]
			if ok {
				fmt.Println("use pool")
				return outCtx, cl.client, nil
			}

			fmt.Println("new client")
			conn, err := grpc.DialContext(ctx, echoEndpoint, grpc.WithDefaultCallOptions(grpc.CallCustomCodec(Codec())), grpc.WithInsecure())//grpc.WithCodec(Codec()))
			p.connects[echoEndpoint] = &clientConn{
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
}
