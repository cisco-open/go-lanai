package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"errors"
	"fmt"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd/lb"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/url"
	"path"
)

type clientDefaults struct {
	selector discovery.InstanceMatcher
	before   []BeforeHook
	after    []AfterHook
}

type client struct {
	defaults   *clientDefaults
	config     *ClientConfig
	discClient discovery.Client
	endpointer Endpointer
	options    []httptransport.ClientOption
}

func NewClient(discClient discovery.Client, opts ...ClientOptions) Client {
	opt := ClientOption{
		ClientConfig: *DefaultConfig(),
		DefaultSelector: discovery.InstanceIsHealthy(),
	}
	for _, f := range opts {
		f(&opt)
	}

	ret := &client {
		config: &opt.ClientConfig,
		discClient: discClient,
		defaults: &clientDefaults{
			selector: opt.DefaultSelector,
			before: opt.DefaultBeforeHooks,
			after: opt.DefaultAfterHooks,
		},
	}
	ret.updateConfig(&opt.ClientConfig)
	return ret
}

func (c *client) WithService(service string, selectors ...discovery.InstanceMatcher) (Client, error) {
	instancer, e := c.discClient.Instancer(service)
	if e != nil {
		return nil, e
	}

	// determine selector
	effectiveSelector := c.defaults.selector
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
	return cp.WithConfig(defaultServiceConfig()), nil
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

	if config.BeforeHooks == nil {
		config.BeforeHooks = c.config.BeforeHooks
	}

	if config.AfterHooks == nil {
		config.AfterHooks = c.config.AfterHooks
	}

	switch {
	case config.MaxRetries < 0:
		config.MaxRetries = 0
	case config.MaxRetries == 0:
		config.MaxRetries = c.config.MaxRetries
	}

	cp := c.shallowCopy()
	cp.updateConfig(config)
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
	ep := lb.RetryWithCallback(c.config.Timeout, b, c.retryCallback(c.config.MaxRetries))

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

// retryCallback is a retry control func.
// It keep trying in case that error is not ErrorTypeResponse and not reached max value
func (c *client) retryCallback(max int) lb.Callback {
	return func(n int, received error) (keepTrying bool, replacement error) {
		return n < max && !errors.Is(received, ErrorTypeResponse), nil
	}
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

func (c *client) makeEndpointFactory(_ context.Context, req *Request, opt *responseOption) EndpointFactory {
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

		cl := httptransport.NewClient(req.Method, uri, req.EncodeFunc, opt.decodeFunc, c.options...)

		return cl.Endpoint(), nil
	}
}

func (c *client) updateConfig(config *ClientConfig) {
	c.config = config
	c.options = make([]httptransport.ClientOption, 0)

	before := append(c.defaults.before, config.BeforeHooks...)
	order.SortStable(before, order.OrderedFirstCompare)
	for _, h := range before {
		c.options = append(c.options, httptransport.ClientBefore(h.RequestFunc()))
	}

	after := append(c.defaults.after, config.AfterHooks...)
	order.SortStable(after, order.OrderedFirstCompare)
	for _, h := range after {
		c.options = append(c.options, httptransport.ClientAfter(h.ResponseFunc()))
	}
}

func (c *client) shallowCopy() *client {
	return &client {
		config: c.config,
		discClient: c.discClient,
		endpointer: c.endpointer,
		options: c.options,
		defaults: c.defaults,
	}
}