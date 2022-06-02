package http

type StatusCode struct {
	Code    uint32
	Details string
}

var (
	StatusOk                  = StatusCode{Code: 200, Details: "OK"}
	StatusCreate              = StatusCode{Code: 201, Details: "Create"}
	StatusNoContent           = StatusCode{Code: 204, Details: "No Content"}
	StatusBadRequest          = StatusCode{Code: 400, Details: "Bad Request"}
	StatusNotFound            = StatusCode{Code: 404, Details: "Not Found"}
	StatusMethodNotAllow      = StatusCode{Code: 405, Details: "Method Not Allow"}
	StatusTooManyRequest      = StatusCode{Code: 429, Details: "Too Many Request"}
	StatusInternalServerError = StatusCode{Code: 500, Details: "Internal Server Error"}
	StatusGatewayTimeout      = StatusCode{Code: 504, Details: "Gateway Timeout"}
)

type Response struct {
	reqId          ReqId
	statusCode     StatusCode
	headers        map[string]string
	body           interface{}
	problemDetails string
}

func ResponseOfId(reqId ReqId) *Response {
	return &Response{
		reqId:   reqId,
		headers: make(map[string]string),
	}
}

func (r *Response) Clone() *Response {
	return &Response{
		reqId:          r.reqId,
		statusCode:     r.statusCode,
		headers:        r.headers,
		body:           r.body,
		problemDetails: r.problemDetails,
	}
}

func (r *Response) AddReqId(reqId ReqId) *Response {
	r.reqId = reqId
	return r
}

func (r *Response) AddStatusCode(statusCode StatusCode) *Response {
	r.statusCode = statusCode
	return r
}

func (r *Response) AddHeader(key, value string) *Response {
	r.headers[key] = value
	return r
}

func (r *Response) AddHeaders(headers map[string]string) *Response {
	for k, v := range headers {
		r.headers[k] = v
	}
	return r
}

func (r *Response) AddBody(body interface{}) *Response {
	r.body = body
	return r
}

func (r *Response) AddProblemDetails(details string) *Response {
	r.problemDetails = details
	return r
}

func (r *Response) ReqId() ReqId {
	return r.reqId
}

func (r *Response) StatusCode() StatusCode {
	return r.statusCode
}

func (r *Response) Headers() map[string]string {
	return r.headers
}

func (r *Response) Header(key string) (string, bool) {
	value, ok := r.headers[key]
	return value, ok
}

func (r *Response) Body() interface{} {
	return r.body
}

func (r *Response) ProblemDetails() string {
	return r.problemDetails
}

// IsSuccess 如果status code为2xx，返回true，否则，返回false
func (r *Response) IsSuccess() bool {
	return r.StatusCode().Code/100 == 2
}
