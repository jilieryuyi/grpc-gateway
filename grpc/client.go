package main

// run: go run client.go api.pb.go
// 简单的grpc客户端实现

import (
	"context"
	"fmt"
	"os"
	"time"
	"google.golang.org/grpc"
	"github.com/jilieryuyi/go-kit-grpc-test-demo/server/src/pb"

	"encoding/json"
)



func MyCodec() grpc.Codec {

	return &myprotoCodec{}
}

type myprotoCodec struct{
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


func (myprotoCodec) Marshal(v interface{}) ([]byte, error) {
	fmt.Printf("myprotoCodec Marshal msg: %v\n", v)
	//return []byte(v.([]byte)), nil
	p2, _ := json.Marshal(v)
	//jtype := []byte("----json----")
	//buf := make([]byte, 0, len(jtype) + len(p2))
	//buf = append(buf, jtype...)
	//buf = append(buf, p2...)
	return p2, nil
}

func (d myprotoCodec) Unmarshal(data []byte, v interface{}) error {
	fmt.Printf("myprotoCodec Unmarshal msg: %v, ,,, %+v\n", string(data), data)
	//return proto.Unmarshal(data, v.(proto.Message))
	return json.Unmarshal(data, v)
}

func (myprotoCodec) String() string {
	fmt.Println("myprotoCodec string")
	return "proto"
}

func (myprotoCodec) Name() string {
	fmt.Println("myprotoCodec string")
	return "proto"
}


func main() {
	grpcAddr       := "127.0.0.1:8081"
	ctx, _ := context.WithTimeout(context.Background(), time.Second)

	opt := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(MyCodec()))

	conn, err := grpc.DialContext(ctx, grpcAddr, grpc.WithInsecure(), opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}
	defer conn.Close()
	svc := pb.NewAddClient(conn)
	req := &pb.SumRequest{
		A:100,
		B:100,
	}
	v, err := svc.Sum(context.Background(), req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stdout, "%d + %d = %d  %v\n", req.A, req.B, v.V, v.Err)
}
