package proxy

import (
	"net/http"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"fmt"
	"encoding/json"
	"strings"
	"github.com/jilieryuyi/grpc-gateway/service"
	"github.com/jilieryuyi/grpc-gateway/proto"
	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
	"time"
	"google.golang.org/grpc/metadata"
	"github.com/jilieryuyi/grpc-gateway/tools"
)

type connection struct {
	conn *grpc.ClientConn
	start int64
}
type MyMux struct {
	conns map[string]*connection
	ctx context.Context
	consulAddress string
	health *api.Health
}

func NewMyMux(ctx context.Context,consulAddress string) *MyMux {
	config := api.DefaultConfig()
	config.Address = consulAddress
	client, err := api.NewClient(config)
	if err != nil {
		log.Panicf("%v", err)
	}
	m := &MyMux{
		ctx : ctx,
		conns: make(map[string]*connection),
		consulAddress:consulAddress,
		health: client.Health(),
	}
	return m
}

type URI struct {
	packageName string
	serviceName string
	version string
	method string
}

func (uri *URI) getServiceName() string {
	st := strings.Split(uri.serviceName, ".")
	serviceName := ""
	for _, v := range st {
		serviceName += strings.ToUpper(v[:1]) + v[1:]
	}
	return fmt.Sprintf("%v.%v", uri.packageName, serviceName)
}

func (uri *URI) getMethod() string {
	return strings.ToUpper(uri.method[:1]) + uri.method[1:]
}

func (p *MyMux) serviceExists(serviceName string) bool {
	cs, _, err := p.health.Service(serviceName, "", true, nil)
	if err != nil {
		log.Errorf("%v", err)
		return false
	}
	return len(cs) > 0
}

func (p *MyMux) getGrpcClient(serviceName string) *connection {
	//clear timeout conn
	// 最长时间缓存nil的client 3秒
	// 防止穿透，一直查询consul
	for key, v := range p.conns {
		if v.conn == nil && time.Now().Unix()-v.start > 3 {
			delete(p.conns, key)
		}
	}
	conn, ok := p.conns[serviceName]
	// 使用连接池
	if ok {
		fmt.Printf("http proxy use pool\n")
		return conn
	}
	conn = &connection{conn:nil, start:time.Now().Unix()}
	p.conns[serviceName] = conn
	if !p.serviceExists(serviceName) {
		return conn
	}
	resl   := service.NewResolver(p.consulAddress)
	rr     := grpc.RoundRobin(resl)
	lb     := grpc.WithBalancer(rr)
	opt    := grpc.WithDefaultCallOptions(grpc.CallCustomCodec(proto.Codec()), grpc.FailFast(false))
	gconn, err := grpc.DialContext(p.ctx, serviceName, opt, grpc.WithInsecure(), lb)
	if err != nil {
		fmt.Printf("http proxy use err nil\n")
		return conn
	}
	conn.conn = gconn
	return conn
}

func (p *MyMux) Close() {
	for _, v := range p.conns {
		if v.conn != nil {
			v.conn.Close()
		}
	}
}

func (p *MyMux) parseURL(url string) *URI {
	// /proto/service.add/v1/sum
	st := strings.Split(url, "/")
	if len(st) < 5 {
		return nil
	}
	return &URI{
		packageName: st[1],
		serviceName: st[2],
		version: st[3],
		method: st[4],
	}
}

// 解析http表单参数，最终转换为需要穿透的grpc参数
func (p *MyMux) parseParams(req *http.Request) map[string]interface{} {
	req.ParseForm()
	// 处理传统意义上表单的参数，这里添加body内传输的json解析支持
	// 解析后的值默认追加到表单内部
	// 支持post、get、json
	params := make(map[string]interface{})
	log.Printf("form:%+v\n\n", req.Form)
	var err error
	for key, v := range req.Form {
		var data map[string]interface{}
		// curl post -d '{"a":"100", "b":"100"}'这种形式过来的数据
		// 会被解析到req.Form的key当中，这时候value是空值
		err = json.Unmarshal([]byte(key), &data)
		if err == nil {
			for kk,vv := range data {
				params[kk] = vv
			}
		} else {
			//常规的表单数据
			if len(v) > 0 {
				fmt.Printf("1 param: %v=>%v\n", key, v[0])
				params[key] = v[0]
			} else {
				fmt.Printf("1 param: %v=>\"\"\n", key)
				params[key] = ""
			}
		}
	}
	// 如果body中有数据，尝试使用json解析
	if req.ContentLength <= 0 {
		return params
	}
	var data map[string]interface{}
	buf := make([]byte, req.ContentLength)
	req.Body.Read(buf)
	fmt.Printf("body: %+v\n" , string(buf))
	err = json.Unmarshal(buf, &data)
	if err != nil || data == nil {
		return params
	}
	for k, dv := range data {
		fmt.Printf("param: %v=>%v\n", k, dv)
		params[k] = dv
	}
	return params
}

// 所有的http请求都会落到这里
func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	// url like:
	// http://localhost:8084/proto/service.add/v1/sum
	// package name is: proto
	// service name is: service.add
	// version is: v1
	// method is: sum
	fmt.Printf("%+v\n", *r)
	fmt.Printf("url: %+v\n", *r.URL)

	uri := p.parseURL(r.URL.Path)
	if uri == nil {
		w.WriteHeader(404)
		w.Write([]byte("url path error, url path must be format by: /{packagename}/{servicename}/{version}/{method}"))
		return
	}
	fmt.Printf("uri: %+v\n", *uri)
	params := p.parseParams(r)
	fmt.Printf("send params: %+v\n", params)

	fullMethod := fmt.Sprintf("/%v/%v", uri.getServiceName(), uri.getMethod())
	fmt.Printf("fullMethod=%s\v", fullMethod)

	conn := p.getGrpcClient(uri.serviceName)
	if conn == nil || conn.conn == nil {
		w.WriteHeader(404)
		w.Write([]byte("connect "+uri.serviceName + " error"))
		return
	}

	header := tools.NewHeader(context.Background())
	//md        := metadata.MD{}
	username  := ""
	password  := ""
	isset     := false

	if r.URL.User != nil {
		password, isset = r.URL.User.Password()
	}

	header.Set("request_uri", r.RequestURI).
	Set("version", uri.version).
	Set("server", "service.gateway").
	Set("method", r.Method).
	Set("opaque", r.URL.Opaque).
	Set("username", username).
	Set("password", password).
	Set("password_set", isset).
	Set("host",r.URL.Host).
	Set("path", r.URL.Path).
	Set("raw_path", r.URL.RawPath).
	Set("force_query", r.URL.ForceQuery).
	Set("raw_query", r.URL.RawQuery).
	Set("fragment", r.URL.Fragment).
	Set("remote_addr", r.RemoteAddr).
	Set("is_form_http","1")
	for key, v := range r.Header {
		header.Set(key, v)
	}
	for key, v := range r.Trailer {
		//grpc的header key只能是小写，大写会报错
		header.Set(key, v)
	}

	// 这里的header发送不过去，待解决
	//ctx:= context.Background()
	ctx := header.ClientContext()//metadata.NewOutgoingContext(ctx, md)
	//这个参数用于接收grpc回传的header和trailer
	var returnHeader  = metadata.MD{}
	var returnTrailer = metadata.MD{}
	opt1 := grpc.Header(&returnHeader)
	opt2 := grpc.Trailer(&returnTrailer)

	//ctx=context.WithValue(ctx, HttpHeader{}, md)
	//err := grpc.Invoke(ctx, fullMethod, params, &out, conn.conn,  grpc.FailFast(false), opt1, opt2)
	var out interface{}
	err := conn.conn.Invoke(ctx, fullMethod, params, &out, opt1, opt2) //grpc.FailFast(false)
	//grpc.SendHeader(ctx, md)
	fmt.Printf("ctx: %+v\n\n", ctx)
	fmt.Printf("return: %+v, error: %+v\n", out, err)
	fmt.Printf("out header: %+v\n", returnHeader)
	fmt.Printf("out trailer: %+v\n", returnTrailer)
	for k, v := range returnHeader {
		if len(v) > 0 {
			for _, sv := range v {
				fmt.Printf("set header: %v=%v\n", k, sv)
				w.Header().Add(k, sv)
			}
		} else {
			w.Header().Set(k, "")
		}
	}
	for k, v := range returnTrailer {
		if len(v) > 0 {
			for _, sv := range v {
				fmt.Printf("set header: %v=%v\n", k, sv)
				w.Header().Add(k, sv)
			}
		} else {
			w.Header().Set(k, "")
		}
	}
	w.Header().Set("Content-Type", "text/html")
	b, err :=json.Marshal(out)
	if err != nil {
		w.Write([]byte("not a json"))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
	return
}