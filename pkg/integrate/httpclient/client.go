// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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

var (
	insecureInstanceMatcher = discovery.InstanceWithTagKV("secure", "false", true)
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

func newClient(discClient discovery.Client, opts ...ClientOptions) Client {
	config := DefaultConfig()
	opt := ClientOption{
		ClientConfig:       *config,
		DefaultSelector:    discovery.InstanceIsHealthy(),
		DefaultBeforeHooks: []BeforeHook{HookRequestLogger(config.Logger, &config.Logging)},
		DefaultAfterHooks:  []AfterHook{HookResponseLogger(config.Logger, &config.Logging)},
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
		return nil, NewNoEndpointFoundError("cannot create client with service name: %v", e)
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
		return nil, NewNoEndpointFoundError("cannot create client with service name: %v", e)
	}

	cp := c.shallowCopy()
	cp.endpointer = endpointer
	return cp.WithConfig(defaultServiceConfig()), nil
}

func (c *client) WithBaseUrl(baseUrl string) (Client, error) {
	endpointer, e := NewSimpleEndpointer(baseUrl)
	if e != nil {
		return nil, NewNoEndpointFoundError("cannot create client with base URL: %v", e)
	}

	cp := c.shallowCopy()
	cp.endpointer = endpointer
	return cp.WithConfig(defaultExtHostConfig()), nil
}

func (c *client) WithConfig(config *ClientConfig) Client {
	mergeConfig(config, c.config)
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
	resp, e := ep(ctx, request)
	if e != nil {
		err = c.translateError(request, e)
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

//nolint:errorlint
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
	return func(instDesp interface{}) (endpoint.Endpoint, error) {
		switch inst := instDesp.(type) {
		case *discovery.Instance:
			return c.endpoint(inst, req, opt)
		case discovery.Instance:
			return c.endpoint(&inst, req, opt)
		case *url.URL:
			return c.simpleEndpoint(inst, req, opt)
		case url.URL:
			return c.simpleEndpoint(&inst, req, opt)
		default:
			return nil, NewInternalError("endpoint is not properly configured: endpoint factory doesn't support instance descriptor %T", instDesp)
		}
	}
}

func (c *client) endpoint(inst *discovery.Instance, req *Request, opt *responseOption) (endpoint.Endpoint, error) {
	ctxPath := ""
	if inst.Meta != nil {
		ctxPath = inst.Meta[discovery.InstanceMetaKeyContextPath]
	}

	scheme := "https"
	if m, e := insecureInstanceMatcher.Matches(inst); m && e == nil {
		scheme = "http"
	}
	uri := &url.URL{
		Scheme: scheme,
		Host: fmt.Sprintf("%s:%d", inst.Address, inst.Port),
		Path: path.Clean(fmt.Sprintf("%s%s", ctxPath, req.Path)),
	}

	cl := httptransport.NewClient(req.Method, uri, effectiveEncodeFunc, opt.decodeFunc, c.options...)

	return cl.Endpoint(), nil
}

func (c *client) simpleEndpoint(baseUrl *url.URL, req *Request, opt *responseOption) (endpoint.Endpoint, error) {
	// make a copy first
	uri := *baseUrl
	// join path
	uri.Path = path.Clean(path.Join(baseUrl.Path, req.Path))
	cl := httptransport.NewClient(req.Method, &uri, effectiveEncodeFunc, opt.decodeFunc, c.options...)

	return cl.Endpoint(), nil
}

func (c *client) updateConfig(config *ClientConfig) {
	c.config = config
	c.options = make([]httptransport.ClientOption, 0)

	if config.HTTPClient != nil {
		c.options = append(c.options, httptransport.SetClient(config.HTTPClient))
	}

	before := append(c.defaults.before, config.BeforeHooks...)
	order.SortStable(before, order.OrderedFirstCompare)
	for _, h := range before {
		if configurable, ok := h.(ConfigurableBeforeHook); ok {
			h = configurable.WithConfig(config)
		}
		c.options = append(c.options, httptransport.ClientBefore(h.RequestFunc()))
	}

	after := append(c.defaults.after, config.AfterHooks...)
	order.SortStable(after, order.OrderedFirstCompare)
	for _, h := range after {
		if configurable, ok := h.(ConfigurableAfterHook); ok {
			h = configurable.WithConfig(config)
		}
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