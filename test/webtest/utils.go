package webtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
)

// CurrentPort utility func that extract current server port from testing context
// Return -1 if not found
func CurrentPort(ctx context.Context) int {
	if v, ok := ctx.Value(ctxKeyInfo).(*serverInfo); ok {
		return v.port
	}
	return -1
}

// CurrentContextPath utility func that extract current server context-path from testing context
// Return DefaultContextPath if not found
func CurrentContextPath(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyInfo).(*serverInfo); ok {
		return v.contextPath
	}
	return DefaultContextPath
}

type RequestOptions func(req *http.Request)

// NewRequest create a new *http.Request based on current execution mode.
// WithRealServer mode:
// 	- Created request have Host, Port and ContextPath set to current TestServer.
// 	- If the given target is relative path, "http" is used. and "context path" is prepended to the given path.
// 	- If the given target is absolute URL, its Host, Port are overridden, and path is kept unchanged
//
// WithMockedServer mode:
//  - the returned request is created by `httptest.NewRequest` and cannot be used by http.DefaultClient.Do()
//	- If the given target is relative path, "http" is used. and "context path" is prepended to the given path.
//	- If the given target is absolute URL, host, port and path are kept unchanged
//
// This function panic if given target is not valid absolute/relative URL or test server is not enabled
func NewRequest(ctx context.Context, method, target string, body io.Reader, opts...RequestOptions) (req *http.Request) {
	tUrl, e := url.Parse(target)
	if e != nil {
		panic(fmt.Sprintf("invalid request target: %v", e))
	}

	info, ok := ctx.Value(ctxKeyInfo).(*serverInfo)
	if !ok {
		panic("invalid use of webtest.NewRequest(). Make sure webtest.WithRealServer() or webtest.WithMockedServer() is in-effect")
	}

	if !tUrl.IsAbs() {
		tUrl.Scheme = "http"
		tUrl.Path = path.Clean(path.Join(info.contextPath, tUrl.Path))
	}

	if ctx.Value(ctxKeyHttpHandler) != nil {
		// WithMockedServer is enabled, we use httptest
		req = httptest.NewRequest(method, tUrl.String(), body).WithContext(ctx)
	} else {
		tUrl.Host = fmt.Sprintf("%s:%d", info.hostname, info.port)
		req, e = http.NewRequestWithContext(ctx, method, tUrl.String(), body)
		if e != nil {
			panic(e)
		}
	}
	for _, fn := range opts {
		fn(req)
	}
	return
}

// Exec execute given request depending on test server mode (real vs mocked)
// returned ExecResult is guaranteed to have non-nil ExecResult.Response if there is no error.
// ExecResult.ResponseRecorder is non-nil if test server mode is WithMockedServer()
// this func might return error if test server mode is WithRealServer()
// Note: don't forget to close the response's body when done with it
//nolint:bodyclose // we don't close body here, whoever using this function should close it when done
func Exec(ctx context.Context, req *http.Request, opts...RequestOptions) (ExecResult, error) {
	for _, fn := range opts {
		fn(req)
	}
	if handler, ok := ctx.Value(ctxKeyHttpHandler).(http.Handler); ok {
		// mocked mode
		rw := httptest.NewRecorder()
		handler.ServeHTTP(rw, req)
		return ExecResult{
			Response:         rw.Result(),
			ResponseRecorder: rw,
		}, nil
	}

	// default to real server mode
	resp, e := http.DefaultClient.Do(req)
	return ExecResult{
		Response: resp,
	}, e
}

// MustExec is same as Exec, but panic instead of returning error
// Note: don't forget to close the response's body when done with it
func MustExec(ctx context.Context, req *http.Request, opts...RequestOptions) ExecResult {
	ret, e := Exec(ctx, req, opts...)
	if e != nil {
		panic(e)
	}
	return ret
}

/*************************
	Options
 *************************/

func WithHeaders(kvs...string) RequestOptions {
	return func(req *http.Request) {
		for i := 0; i < len(kvs); i+=2 {
			if i + 1 < len(kvs) {
				req.Header.Add(kvs[i], kvs[i+1])
			} else {
				req.Header.Add(kvs[i], "")
			}
		}
	}
}

func WithQueries(kvs ...string) RequestOptions {
	return func(req *http.Request) {
		q := req.URL.Query()
		for i := 0; i < len(kvs); i += 2 {
			if i+1 < len(kvs) {
				q.Add(kvs[i], kvs[i+1])
			} else {
				q.Add(kvs[i], "")
			}
		}
		req.URL.RawQuery = q.Encode()
	}
}

func WithCookies(resp *http.Response) RequestOptions {
	cookies := resp.Cookies()
	kvs := make([]string, len(cookies)*2)
	for i := range cookies {
		kvs[i*2] = "Cookie"
		kvs[i*2+1] = cookies[i].String()
	}
	return WithHeaders(kvs...)
}

/*************************
	Custom Context
 *************************/

type infoCtxKey struct{}

var ctxKeyInfo = infoCtxKey{}

type httpHandlerCtxKey struct{}

var ctxKeyHttpHandler = httpHandlerCtxKey{}

type serverInfo struct {
	hostname    string
	port        int
	contextPath string
}

type webTestContext struct {
	context.Context
	info    *serverInfo
	handler http.Handler
}

func newWebTestContext(parent context.Context, info *serverInfo, handler http.Handler) context.Context {
	return &webTestContext{
		Context: parent,
		info:    info,
		handler: handler,
	}
}

func (c *webTestContext) Value(key interface{}) interface{} {
	switch {
	case key == ctxKeyInfo && c.info != nil:
		return c.info
	case key == ctxKeyHttpHandler && c.handler != nil:
		return c.handler
	}
	return c.Context.Value(key)
}
