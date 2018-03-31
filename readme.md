grpc proxy gateway
------------------
use consul for register and get services
support
-------
````
grpc service -> gateway -> grpc service
http         -> gateway -> grpc service
````

grpc load balancing

使用consul进行服务注册和服务发现<br/>
同时支持http、grpc无依赖proto穿透<br/>
支持gateway负载均衡和服务负载均衡<br/>
<br/>
tools目录下面包含了用于mac下面测试的consul和zipkin，如果在mac下测试可直接启用<br/>
其他环境可以用做参考<br/>
<br/>
tools/header.go 用于解析grpc header<br/>
services目录为测试的grpc服务端，使用gokit构建<br/>
service目录下为consul服务注册已经服务发现相关实现<br/>
protocol为proto相关实现，services和client依赖此实现<br/>
proto为proto协议相关实现<br/>
gateway为网关实现<br/>
examples为相关示例<br/>
client为使用gokit搭建的grpc示例客户端<br/>
<br/>