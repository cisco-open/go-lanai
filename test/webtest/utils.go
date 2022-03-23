package webtest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
)

// CurrentPort utility func that extract current server port from testing context
// Return -1 if not found
func CurrentPort(ctx context.Context) int {
	if v, ok := ctx.Value(ctxKeyAddr).(*addr); ok {
		return v.port
	}
	return -1
}

// CurrentContextPath utility func that extract current server context-path from testing context
// Return DefaultContextPath if not found
func CurrentContextPath(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyAddr).(*addr); ok {
		return v.contextPath
	}
	return DefaultContextPath
}

// NewRequest create a new *http.Request with Host, Port and ContextPath are set to current TestServer.
// If the given target is relative path, "http" is used. and "context path" is prepended to the given path.
// If the given target is absolute URL, its Host, Port are overridden, and path is kept unchanged
// This function panic if given target is not valid absolute/relative URL or it's not in real test server mode
func NewRequest(ctx context.Context, method, target string, body io.Reader) *http.Request {
	tUrl, e := url.Parse(target)
	if e != nil {
		panic(fmt.Sprintf("invalid request target: %v", e))
	}

	// we  panic here if addr is not found
	if v, ok := ctx.Value(ctxKeyAddr).(*addr); ok {
		tUrl.Host = fmt.Sprintf("%s:%d", v.hostname, v.port)
		if !tUrl.IsAbs() {
			tUrl.Scheme = "http"
			tUrl.Path = path.Clean(path.Join(v.contextPath, tUrl.Path))
		}
	} else {
		panic("invalid use of webtest.NewRequest(). Make sure webtest.WithRealTestServer() is in-effect")
	}
	req, e := http.NewRequest(method, tUrl.String(), body)
	if e != nil {
		panic(e)
	}
	return req
}


/*************************
	Custom Context
 *************************/

type addrCtxKey struct{}

var ctxKeyAddr = addrCtxKey{}

type addr struct {
	hostname    string
	port        int
	contextPath string
}

type addrAwareTestContext struct {
	context.Context
	addr *addr
}

func contestWithAddr(parent context.Context, addr addr) context.Context {
	return &addrAwareTestContext{
		Context: parent,
		addr:    &addr,
	}
}

func (c *addrAwareTestContext) Value(key interface{}) interface{} {
	switch {
	case key == ctxKeyAddr && c.addr != nil:
		return c.addr
	}
	return c.Context.Value(key)
}
