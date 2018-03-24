package proto

import (
	"fmt"
	"google.golang.org/grpc"
	"encoding/json"
)


func Codec() grpc.Codec {
	return &myprotoCodec{}
}

type myprotoCodec struct{}


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

