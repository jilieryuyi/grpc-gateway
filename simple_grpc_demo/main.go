package main

// run: go run main.go api.pb.go

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"github.com/oklog/oklog/pkg/group"
	"google.golang.org/grpc"
	kitgrpc "github.com/go-kit/kit/transport/grpc"
	oldcontext "golang.org/x/net/context"
)

type api struct {
	AddClient
}

func (s *api) Sum(ctx oldcontext.Context, req *SumRequest) (*SumReply, error) {
	rep := &SumReply{
		V:req.A + req.B,
	}
	return rep, nil
}

func main() {
	var g group.Group
	grpcListener, err := net.Listen("tcp", "127.0.0.1:9996")
	if err != nil {
		os.Exit(1)
	}
	g.Add(func() error {
		baseServer := grpc.NewServer(grpc.UnaryInterceptor(kitgrpc.Interceptor))
		RegisterAddServer(baseServer, &api{})
		return baseServer.Serve(grpcListener)
	}, func(error) {
		grpcListener.Close()
	})
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
	g.Run()
}
