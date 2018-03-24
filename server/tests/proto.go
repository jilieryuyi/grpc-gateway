package main

import (
	"fmt"
	"os"
	"github.com/RangelReale/fproto"
)

func main() {
	file, err := os.Open("/Users/yuyi/Code/go/gokit/grpc/server/src/pb/addsvc.proto")
	if err != nil {
		return// nil, err
	}
	defer file.Close()

	protofile, err := fproto.Parse(file)
	if err != nil {
		return// nil, err
	}

	fmt.Printf("%+v\n\n", *protofile)
	for key, v := range protofile.Messages {
		fmt.Printf("%v=>%v\n", key, v)
		fmt.Println("name=",v.Name)
		fmt.Printf("%+v\n",*v.Comment)
		//fmt.Printf("%+v\n",*v.Enums)
		for _, e := range v.Enums {
			fmt.Printf("===>%+v\n", *e)
		}
		for _, e := range v.Extensions {
			fmt.Printf("===>%+v\n", *e)
		}
		for _, e := range v.Fields {
			fmt.Printf("Fields===>%+v\n", e)
		}
		fmt.Printf("\n\n")
	}
	fmt.Println("\n\n====\n\n")
	for key, v := range protofile.Services {
		fmt.Printf("%v=>%v\n", key, *v)
		fmt.Println(v.Name)
		for kr, vr := range v.RPCs {
			fmt.Printf("%v, %v\n", kr, vr)
			for _, ov := range vr.Options {
				fmt.Printf("Options=>%+v\n", *ov)
			}
		}
	}
}