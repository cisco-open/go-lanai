package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/url"
	"path"
	"time"
)

type ClientOptions func(opt *ClientOption)

type ClientOption struct {
	ClientConfig
}

type ClientConfig struct {
	MaxRetries int
	Timeout    time.Duration
	Logger     log.ContextualLogger
	Verbose    bool
}

type client struct {
	config *ClientConfig
	discClient discovery.Client
	endpointer Endpointer
	//request *Request
	//response *Response
}

func NewClient(discClient discovery.Client, opts ...ClientOptions) Client {
	opt := ClientOption{
		ClientConfig: ClientConfig{
			MaxRetries: 3,
			Timeout: 2 * time.Minute,
			Logger: logger,
		},
	}
	for _, f := range opts {
		f(&opt)
	}

	return &client {
		config: &opt.ClientConfig,
		discClient: discClient,
	}
}

func (c *client) WithService(service string, selectors ...discovery.InstanceMatcher) (Client, error) {
	instancer, e := c.discClient.Instancer(service)
	if e != nil {
		return nil, e
	}

	// determine selector
	effectiveSelector := discovery.InstanceIsHealthy()
	if len(selectors) != 0 {
		matchers := make([]matcher.Matcher, len(selectors))
		for i, m := range selectors {
			matchers[i] = m
		}
		effectiveSelector = matcher.And(matchers[0], matchers[1:]...)
	}

	endpointer, e := NewKitEndpointer(instancer, func(opts *EndpointerOption) {
		opts.ServiceName = service
		opts.Selector = effectiveSelector
		opts.InvalidateOnError = true
		opts.Logger = c.config.Logger
	})
	if e != nil {
		return nil, e
	}

	cp := c.shallowCopy()
	cp.endpointer = endpointer
	return cp, nil
}

func (c *client) WithBaseUrl(baseUrl string) (Client, error) {
	// TODO implement this
	return nil, fmt.Errorf("not implemented yet")
}

func (c *client) WithConfig(config *ClientConfig) Client {
	if config.Logger == nil {
		config.Logger = c.config.Logger
	}

	if config.Timeout <= 0 {
		config.Timeout = c.config.Timeout
	}

	cp := c.shallowCopy()
	cp.config = config
	return cp
}

func (c *client) Execute(ctx context.Context, request *Request, opts ...ResponseOptions) (ret *Response, err error) {
	// apply options
	opt := responseOption{}
	for _, f := range opts {
		f(&opt)
	}

	// apply fallback options
	fallbackResponseOptions(&opt)

	// create endpoint
	epConfig := EndpointerConfig{
		EndpointFactory: c.makeEndpointFactory(ctx, request, &opt),
	}
	b := lb.NewRoundRobin(c.endpointer.WithConfig(&epConfig))
	ep := lb.Retry(c.config.MaxRetries, c.config.Timeout, b)

	// execute endpoint and handle error
	resp, e := ep(ctx, request.Body)
	if e != nil {
		err = c.translateError(request, e)
	}

	// verbose log
	if c.config.Verbose {
		// TODO verbose log
	}

	// return result
	switch v := resp.(type) {
	case *Response:
		ret = v
	case Response:
		ret = &v
	default:
		if err == nil {
			err = NewInternalError(fmt.Errorf("expected a *Response, but HTTP response decode function returned %T", resp))
		}
	}
	return
}

func (c *client) translateError(req *Request, err error) *Error {
	switch retry := err.(type) {
	case lb.RetryError:
		err = retry.Final
	case *lb.RetryError:
		err = retry.Final
	}

	switch {
	case errors.Is(err, ErrorCategoryHttpClient):
		return err.(*Error)
	case err == context.Canceled:
		e := fmt.Errorf("remote HTTP call [%s] %s timed out after %v", req.Method, req.Path, c.config.Timeout)
		return NewServerTimeoutError(e)
	case err == lb.ErrNoEndpoints:
		e := fmt.Errorf("remote HTTP call [%s] %s: no endpoints available", req.Method, req.Path)
		return NewNoEndpointFoundError(e)
	default:
		e := fmt.Errorf("uncategrized remote HTTP call [%s] %s error: %v", req.Method, req.Path, err)
		return NewInternalError(e)
	}
}

func (c *client) makeEndpointFactory(ctx context.Context, req *Request, opt *responseOption) EndpointFactory {
	return func(inst *discovery.Instance) (endpoint.Endpoint, error) {
		ctxPath := ""
		if inst.Meta != nil {
			ctxPath = inst.Meta[discovery.InstanceMetaKeyContextPath]
		}

		// TODO choose http or https based on tag "secure"
		uri := &url.URL{
			Scheme: "http",
			Host: fmt.Sprintf("%s:%d", inst.Address, inst.Port),
			Path: path.Clean(fmt.Sprintf("%s%s", ctxPath, req.Path)),
		}

		// TODO install before/after hooks
		cl := httptransport.NewClient(req.Method, uri, req.EncodeFunc, opt.decodeFunc, )

		return cl.Endpoint(), nil
	}
}

func (c *client) shallowCopy() *client {
	return &client {
		config: c.config,
		discClient: c.discClient,
		endpointer: c.endpointer,
	}
}