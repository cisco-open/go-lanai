package httpclient

import (
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
)

type RequestOptions func(r *Request)

type Request struct {
	Path string
	Method string
	Params map[string]string
	Headers http.Header
	Body interface{}
	EncodeFunc httptransport.EncodeRequestFunc
}

func NewRequest(path, method string, opts ...RequestOptions) *Request {
	r := Request{
		Path: path,
		Method: method,
		Params: map[string]string{},
		Headers: http.Header{},
		EncodeFunc: httptransport.EncodeJSONRequest,
	}
	for _, f := range opts {
		f(&r)
	}
	return &r
}

func WithoutHeader(key string) RequestOptions {
	switch {
	case key == "":
		return func(r *Request) {}
	default:
		return func(r *Request) {
			r.Headers.Del(key)
		}
	}
}

func WithHeader(key, value string) RequestOptions {
	switch {
	case key == "" || value == "":
		return func(r *Request) {}
	default:
		return func(r *Request) {
			r.Headers.Add(key, value)
		}
	}
}

func WithParam(key, value string) RequestOptions {
	switch {
	case key == "":
		return func(r *Request) {}
	case value == "":
		return func(r *Request) {
			delete(r.Params, key)
		}
	default:
		return func(r *Request) {
			r.Params[key] = value
		}
	}
}

func WithBody(body interface{}) RequestOptions {
	return func(r *Request) {
		r.Body = body
	}
}

func WithRequestEncodeFunc(enc httptransport.EncodeRequestFunc) RequestOptions {
	return func(r *Request) {
		r.EncodeFunc = enc
	}
}
