build proto
-----------
````
protoc api.proto --go_out=plugins=grpc:.
````

run test
--------
````
##start grpc server
go run main.go api.pb.go
##run grpc client
go run client.go api.pb.go
````
