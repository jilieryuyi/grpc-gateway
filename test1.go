package main

import "context"
import "errors"
import "strings"
import "github.com/go-kit/kit/endpoint"
import httptransport "github.com/go-kit/kit/transport/http"
import "encoding/json"
import "net/http"
import "github.com/go-kit/kit/log"
import (
	"os"
	"fmt"
)

type StringService interface {
	Uppercase(context.Context, string) (string, error)
	Count(context.Context, string) int
}

var ErrEmpty = errors.New("empty string")

type stringService struct {
}

//真正处理Uppercase业务
func (stringService) Uppercase(_ context.Context, s string) (string, error) {
	if s == "" {
		return "", ErrEmpty
	}

	return strings.ToUpper(s), nil
}

//真正处理Count业务
func (stringService) Count(_ context.Context, s string) int {
	return len(s)
}

var _ StringService = &stringService{}




//rpc协议定义
type uppercaseRequest struct {
	S string `json:"s"`
}
type uppercaseResponse struct {
	V   string `json:"v"`
	Err string `json:"err,omitempty"`
}
type countRequest struct {
	S string `json:"s"`
}
type countResponse struct {
	V int `json:"v"`
}


func makeUppercaseEndpoint(svc StringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(uppercaseRequest)
		v, err := svc.Uppercase(ctx, req.S)
		if err != nil {
			return uppercaseResponse{v, err.Error()}, nil
		}
		return uppercaseResponse{v, ""}, nil
	}
}

func makeCountEndpoint(svc stringService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(countRequest)
		v := svc.Count(ctx, req.S)
		return countResponse{v}, nil
	}
}

func decodeUpperCaseRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request uppercaseRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func decodeCountRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var request countRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeResponse(_ context.Context, w http.ResponseWriter, respon interface{}) error {
	return json.NewEncoder(w).Encode(respon)
}


type Middleware func(endpoint.Endpoint) endpoint.Endpoint
func loggingMiddleware(logger log.Logger) endpoint.Middleware {

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			logger.Log("msg", "calling endpoint")
			defer logger.Log("msg", "callend endpoing")
			fmt.Println(ctx, request)
			return next(ctx, request)
		}
	}
}

//curl -XPOST -d'{"s":"Hello"}' http://localhost:8082/count
//{"v":5}
//curl -XPOST -d'{"s":"Hello"}' http://localhost:8082/uppercase
//{"v":"HELLO"}

func main() {
	logger := log.NewLogfmtLogger(os.Stderr)
	mid := loggingMiddleware(log.With(logger, "LOGMETHOD", "uppercase"))

	svc := stringService{}

	uppercaseHandler := httptransport.NewServer(
		mid(makeUppercaseEndpoint(svc)),
		decodeUpperCaseRequest,
		encodeResponse,
	)

	countHandler := httptransport.NewServer(
		mid(makeCountEndpoint(svc)),
		decodeCountRequest,
		encodeResponse,
	)
	http.Handle("/uppercase", uppercaseHandler)
	http.Handle("/count", countHandler)
	http.ListenAndServe(":8082", nil)
}
