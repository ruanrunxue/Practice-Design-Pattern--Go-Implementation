package http

import "math/rand"

type requestBuilder struct {
	req *Request
}

// NewRequestBuilder 普通Builder工厂方法，新创建一个Request对象
func NewRequestBuilder() *requestBuilder {
	return &requestBuilder{req: EmptyRequest()}
}

// NewRequestBuilderCopyFrom 复制已有的Request对象
func NewRequestBuilderCopyFrom(req *Request) *requestBuilder {
	reqId := rand.Uint32() % 10000
	replica := &Request{
		reqId:       ReqId(reqId),
		method:      req.method,
		uri:         req.uri,
		queryParams: req.queryParams,
		headers:     req.headers,
		body:        req.body,
	}
	return &requestBuilder{req: replica}
}

func (r *requestBuilder) AddMethod(method Method) *requestBuilder {
	r.req.method = method
	return r
}

func (r *requestBuilder) AddUri(uri Uri) *requestBuilder {
	r.req.uri = uri
	return r
}

func (r *requestBuilder) AddQueryParam(key, value string) *requestBuilder {
	r.req.queryParams[key] = value
	return r
}

func (r *requestBuilder) AddQueryParams(params map[string]string) *requestBuilder {
	for k, v := range params {
		r.req.queryParams[k] = v
	}
	return r
}

func (r *requestBuilder) AddHeader(key, value string) *requestBuilder {
	r.req.headers[key] = value
	return r
}

func (r *requestBuilder) AddHeaders(headers map[string]string) *requestBuilder {
	for k, v := range headers {
		r.req.headers[k] = v
	}
	return r
}

func (r *requestBuilder) AddBody(body interface{}) *requestBuilder {
	r.req.body = body
	return r
}

func (r *requestBuilder) Builder() *Request {
	return r.req
}
