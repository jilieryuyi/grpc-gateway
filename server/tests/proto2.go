package main

import (
	"fmt"
	"os"

	"github.com/emicklei/proto"
)

func main() {
	reader, _ := os.Open("/Users/yuyi/Code/go/gokit/grpc/server/src/pb/addsvc.proto")
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, _ := parser.Parse()

	proto.Walk(definition,
		proto.WithService(handleService),
		proto.WithMessage(handleMessage))
}

func handleService(s *proto.Service) {
	fmt.Println(s.Name)
}

func handleMessage(m *proto.Message) {
	fmt.Println(m.Name)
}