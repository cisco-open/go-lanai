package httpclient

import (
	"context"
	"encoding/base64"
	"fmt"
	httptransport "github.com/go-kit/kit/transport/http"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type RequestOptions func(r *Request)

// Request is wraps all information about the request
type Request struct {
	Path       string
	Method     string
	Params     map[string]string
	Headers    http.Header
	Body       interface{}
	EncodeFunc httptransport.EncodeRequestFunc
}

func NewRequest(path, method string, opts ...RequestOptions) *Request {
	r := Request{
		Path:       path,
		Method:     method,
		Params:     map[string]string{},
		Headers:    http.Header{},
		EncodeFunc: httptransport.EncodeJSONRequest,
	}
	for _, f := range opts {
		f(&r)
	}
	return &r
}

func effectiveEncodeFunc(ctx context.Context, req *http.Request, val interface{}) error {
	var r *Request
	switch v := val.(type) {
	case *Request:
		r = v
	case Request:
		r = &v
	default:
		return NewRequestSerializationError(fmt.Errorf("request encoder expects *Request but got %T", val))
	}

	// set headers
	for k := range r.Headers {
		req.Header.Set(k, r.Headers.Get(k))
	}

	// set params
	applyParams(req, r.Params)

	return r.EncodeFunc(ctx, req, r.Body)
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

func WithBasicAuth(username, password string) RequestOptions {
	raw := username + ":" + password
	b64 := base64.StdEncoding.EncodeToString([]byte(raw))
	auth := "Basic " + b64
	return WithHeader(HeaderAuthorization, auth)
}

func WithUrlEncodedBody(body url.Values) RequestOptions {
	return func(r *Request) {
		r.Headers.Set(HeaderContentType, MediaTypeFormUrlEncoded)
		r.Body = body
		r.EncodeFunc = urlEncodedBodyEncoder
	}
}

func urlEncodedBodyEncoder(_ context.Context, r *http.Request, v interface{}) error {
	values, ok := v.(url.Values)
	if !ok {
		return NewRequestSerializationError(fmt.Errorf("www-form-urlencoded body expects url.Values but got %T", v))
	}
	reader := strings.NewReader(values.Encode())
	r.Body = ioutil.NopCloser(reader)
	return nil
}

func applyParams(req *http.Request, params map[string]string) {
	if len(params) == 0 {
		return
	}

	queries := make([]string, len(params))
	i := 0
	for k, v := range params {
		queries[i] = k + "=" + url.QueryEscape(v)
		i++
	}
	req.URL.RawQuery = strings.Join(queries, "&")
}



