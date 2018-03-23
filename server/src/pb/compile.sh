#!/usr/bin/env sh

# Install proto3 from source
#  brew install autoconf automake libtool
#  git clone https://github.com/google/protobuf
#  ./autogen.sh ; ./configure ; make ; make install
#
# Update protoc Go bindings via
#  go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
#
# See also
#  https://github.com/grpc/grpc-go/tree/master/examples

##protoc addsvc.proto --go_out=plugins=grpc:.
protoc -I/usr/local/include -I. -I/Users/yuyi/go/src -I/Users/yuyi/go/src/github.com/googleapis/googleapis --go_out=plugins=grpc:. *.proto
protoc -I/usr/local/include -I. -I/Users/yuyi/go/src -I/Users/yuyi/go/src/github.com/googleapis/googleapis --grpc-gateway_out=logtostderr=true:. *.proto
