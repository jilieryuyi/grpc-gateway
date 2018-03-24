package proxy

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"encoding/json"
	"google.golang.org/grpc/encoding"
)

// Codec returns a proxying grpc.Codec with the default protobuf codec as parent.
//
// See CodecWithParent.
func Codec() grpc.Codec {
	return CodecWithParent(&protoCodec{})
}

// CodecWithParent returns a proxying grpc.Codec with a user provided codec as parent.
//
// This codec is *crucial* to the functioning of the proxy. It allows the proxy server to be oblivious
// to the schema of the forwarded messages. It basically treats a gRPC message frame as raw bytes.
// However, if the server handler, or the client caller are not proxy-internal functions it will fall back
// to trying to decode the message using a fallback codec.
func CodecWithParent(fallback grpc.Codec) grpc.Codec {
	return &rawCodec{fallback}
}

type rawCodec struct {
	parentCodec grpc.Codec//grpc.Codec
}

type frame struct {
	payload []byte
}

func (c *rawCodec) Marshal(v interface{}) ([]byte, error) {
	fmt.Printf("rawCodec Marshal\n")
	out, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Marshal(v)
	}
	return out.payload, nil

}

func (c *rawCodec) Unmarshal(data []byte, v interface{}) error {
	fmt.Printf("rawCodec Unmarshal\n")
	dst, ok := v.(*frame)
	if !ok {
		return c.parentCodec.Unmarshal(data, v)
	}
	dst.payload = data
	return nil
}

func (c *rawCodec) Name() string {
	return fmt.Sprintf("proxy>%s", c.parentCodec.String())
}

func (c *rawCodec) String() string {
	return fmt.Sprintf("proxy>%s", c.parentCodec.String())
}

// protoCodec is a Codec implementation with protobuf. It is the default rawCodec for gRPC.
type protoCodec struct{}

func (protoCodec) Marshal(v interface{}) ([]byte, error) {
	return proto.Marshal(v.(proto.Message))
}

func (protoCodec) Unmarshal(data []byte, v interface{}) error {
	return proto.Unmarshal(data, v.(proto.Message))
}

func (protoCodec) String() string {
	return "proto"
}




func MyCodec(DefauleCode encoding.Codec) grpc.Codec {

	return &myprotoCodec{DefauleCode}
}

type myprotoCodec struct{
	DefauleCode encoding.Codec
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
