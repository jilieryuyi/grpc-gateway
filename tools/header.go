package tools

import (
	"google.golang.org/grpc/metadata"
	"strings"
	"golang.org/x/net/context"
	"strconv"
)

type Header struct {
	//request_uri string
	//version         string
	//server          string
	//method          string
	//opaque          string
	//username   string
	//password   string
	//password_set bool
	//host            string
	//path            string
	//raw_path         string
	//force_query      string
	//raw_query        string
	//fragment        string
	//remote_addr      string
	//is_form_http bool
	ctx context.Context

	headers  map[string]string
	//trailers map[string]string
	cookies  map[string] string
}



// 注意：所有的key最终会被转换为小写
// 示例
/*
import "github.com/jilieryuyi/grpc-gateway/tools/header"
//这里的ctx来源于serverimp.go的接口上下文ctx context.Context
h := header.NewHeader(ctx)
appid := h.GetCookie("appid")
*/

func NewHeader(ctx context.Context) *Header{
	header := &Header{
		ctx : ctx,
		headers : make(map[string] string),
		cookies : make(map[string] string),
		//trailers: make(map[string] string),
	}
	header.parse()
	return header
}

func (h *Header) parse() {
	md, ok := metadata.FromIncomingContext(h.ctx)
	if ok {
		for key, value := range md {
			if len(value) > 0 {
				h.headers[key] = value[0]
			} else {
				h.headers[key] = ""
			}
		}
	}
	cookies, ok := md["cookie"]
	if ok {
		if len(cookies) > 0 {
			c := strings.Split(cookies[0], ";")
			for _, iv := range c {
				//log.Debugf("cookie==%s", iv)
				iv = strings.Trim(iv, " ")
				it := strings.Split(iv, "=")
				if len(it) >= 2 {
					k := strings.ToLower(strings.Trim(it[0], " "))
					h.cookies[k] = strings.Trim(it[1], " ")
				}
			}
		}
	}
}

// 如果返回空字符串，说明key不存在
func (h *Header) GetHeader(key string) string {
	v, ok := h.headers[strings.ToLower(key)]
	if ok {
		return v
	}
	return ""
}

// 所有的key都是小写
func (h *Header) GetHeaders() map[string] string {
	return h.headers
}

// 如果返回空字符串，说明cookie不存在
func (h *Header) GetCookie(key string) string {
	v, ok := h.cookies[strings.ToLower(key)]
	if ok {
		return v
	}
	return ""
}

// 所有的key都是小写
func (h *Header) GetCookies() map[string] string {
	return h.cookies
}

func (h *Header) GetCookieInt64(key string) int64 {
	v, ok := h.cookies[strings.ToLower(key)]
	if ok {
		d, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return 0
		}
		return d
	}
	return 0
}
