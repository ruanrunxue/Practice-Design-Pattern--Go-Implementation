package http

import (
	"math/rand"
	"strings"
)

type Method uint8

const (
	GET Method = iota + 1
	POST
	PUT
	DELETE
)

type Uri string

func (u Uri) Contains(other Uri) bool {
	return strings.Contains(string(u), string(other))
}

type ReqId uint32

type Request struct {
	reqId       ReqId
	method      Method
	uri         Uri
	queryParams map[string]string
	headers     map[string]string
	body        interface{}
}

func EmptyRequest() *Request {
	reqId := rand.Uint32() % 10000
	return &Request{
		reqId:       ReqId(reqId),
		uri:         "",
		queryParams: make(map[string]string),
		headers:     make(map[string]string),
	}
}

func (r *Request) IsInValid() bool {
	return r.method < 1 || r.method > 4 || r.uri == ""
}

func (r *Request) AddMethod(method Method) *Request {
	r.method = method
	return r
}

func (r *Request) AddUri(uri Uri) *Request {
	r.uri = uri
	return r
}

func (r *Request) AddQueryParam(key, value string) *Request {
	r.queryParams[key] = value
	return r
}

func (r *Request) AddQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.queryParams[k] = v
	}
	return r
}

func (r *Request) AddHeader(key, value string) *Request {
	r.headers[key] = value
	return r
}

func (r *Request) AddHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

func (r *Request) AddBody(body interface{}) *Request {
	r.body = body
	return r
}

func (r *Request) ReqId() ReqId {
	return r.reqId
}

func (r *Request) Method() Method {
	return r.method
}

func (r *Request) Uri() Uri {
	return r.uri
}

func (r *Request) QueryParams() map[string]string {
	return r.queryParams
}

func (r *Request) QueryParam(key string) (string, bool) {
	value, ok := r.queryParams[key]
	return value, ok
}

func (r *Request) Headers() map[string]string {
	return r.headers
}

func (r *Request) Header(key string) (string, bool) {
	value, ok := r.headers[key]
	return value, ok
}

func (r *Request) Body() interface{} {
	return r.body
}
