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
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/pkg/utils/order"
	"net/url"
	"time"
)

var (
	httpInstanceMatcher = discovery.InstanceWithTagKV("secure", "false", true).
				Or(discovery.InstanceWithTagKV("insecure", "true", true)).
				Or(discovery.InstanceWithMetaKV("scheme", "http"))
	httpsInstanceMatcher = discovery.InstanceWithTagKV("secure", "true", true).
		Or(discovery.InstanceWithTagKV("insecure", "false", true)).
		Or(discovery.InstanceWithMetaKV("scheme", "https"))
	supportedSchemes = utils.NewStringSet("http", "https")
)

type clientDefaults struct {
	selector discovery.InstanceMatcher
	before   []BeforeHook
	after    []AfterHook
}

type client struct {
	defaults *clientDefaults
	config   *ClientConfig
	sdClient discovery.Client
	before   []BeforeHook
	after    []AfterHook
	resolver TargetResolver
}

func NewClient(opts ...ClientOptions) Client {
	config := DefaultConfig()
	opt := ClientOption{
		ClientConfig:       *config,
		DefaultSelector:    discovery.InstanceIsHealthy(),
		DefaultBeforeHooks: []BeforeHook{HookRequestLogger(config)},
		DefaultAfterHooks:  []AfterHook{HookResponseLogger(config)},
	}
	for _, f := range opts {
		f(&opt)
	}

	ret := &client{
		config:   &opt.ClientConfig,
		sdClient: opt.SDClient,
		defaults: &clientDefaults{
			selector: opt.DefaultSelector,
			before:   opt.DefaultBeforeHooks,
			after:    opt.DefaultAfterHooks,
		},
	}
	ret.updateConfig(&opt.ClientConfig)
	return ret
}

func (c *client) WithService(service string, opts ...SDOptions) (Client, error) {
	if c.sdClient == nil {
		return nil, NewNoEndpointFoundError("cannot create client with service name: service discovery client is not configured")
	}

	instancer, e := c.sdClient.Instancer(service)
	if e != nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot create client with service name: %s", service), e)
	}

	defaultOpts := func(opts *SDOption) {
		opts.Selector = c.defaults.selector
		opts.InvalidateOnError = true
	}
	opts = append([]SDOptions{defaultOpts}, opts...)
	targetResolver, e := NewSDTargetResolver(instancer, opts...)
	if e != nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot create client with service name: %s", service), e)
	}

	cp := c.shallowCopy()
	cp.resolver = targetResolver
	return cp.WithConfig(defaultServiceConfig()), nil
}

func (c *client) WithBaseUrl(baseUrl string) (Client, error) {
	endpointer, e := NewStaticTargetResolver(baseUrl)
	if e != nil {
		return nil, NewNoEndpointFoundError(fmt.Errorf("cannot create client with base URL: %s", baseUrl), e)
	}

	cp := c.shallowCopy()
	cp.resolver = endpointer
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

	// execute
	executor := c.executor(request, c.resolver, opt.decodeFunc)
	retryCB := c.config.RetryCallback
	if retryCB == nil {
		retryCB = c.retryCallback()
	}
	resp, e := executor.Try(ctx, c.config.Timeout, retryCB)
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
func (c *client) retryCallback() RetryCallback {
	return func(n int, rs interface{}, err error) (bool, time.Duration) {
		return n < c.config.MaxRetries && !errors.Is(err, ErrorTypeResponse), c.config.RetryBackoff
	}
}

func (c *client) translateError(req *Request, err error) (ret *Error) {
	switch {
	case errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded):
		e := fmt.Errorf("remote HTTP call [%s] %s timed out after %v", req.Method, req.Path, c.config.Timeout)
		return NewServerTimeoutError(e)
	case errors.Is(err, ErrorSubTypeDiscovery):
		errors.As(err, &ret)
		return ret.WithMessage("remote HTTP call [%s] %s: no endpoints available", req.Method, req.Path)
	case errors.Is(err, ErrorCategoryHttpClient):
		errors.As(err, &ret)
		return
	default:
		e := fmt.Errorf("uncategrized remote HTTP call [%s] %s error: %v", req.Method, req.Path, err)
		return NewInternalError(e)
	}
}

func (c *client) updateConfig(config *ClientConfig) {
	c.config = config

	c.before = make([]BeforeHook, len(c.defaults.before)+len(config.BeforeHooks))
	copy(c.before, c.defaults.before)
	copy(c.before[len(c.defaults.before):], config.BeforeHooks)
	for i := range c.before {
		if configurable, ok := c.before[i].(ConfigurableBeforeHook); ok {
			c.before[i] = configurable.WithConfig(config)
		}
	}
	order.SortStable(c.before, order.OrderedFirstCompare)

	c.after = make([]AfterHook, len(c.defaults.after)+len(config.AfterHooks))
	copy(c.after, c.defaults.after)
	copy(c.after[len(c.defaults.after):], config.AfterHooks)
	for i := range c.after {
		if configurable, ok := c.after[i].(ConfigurableAfterHook); ok {
			c.after[i] = configurable.WithConfig(config)
		}
	}
	order.SortStable(c.after, order.OrderedFirstCompare)
}

func (c *client) shallowCopy() *client {
	cpy := *c
	return &cpy
}

func (c *client) executor(request *Request, resolver TargetResolver, dec DecodeResponseFunc) Retryable {
	return func(ctx context.Context) (interface{}, error) {
		target, e := url.Parse(request.Path)
		// only need to resolve the target if the request.Path is not absolute
		if e != nil || !supportedSchemes.Has(target.Scheme) {
			target, e = resolver.Resolve(ctx, request)
			if e != nil {
				return nil, e
			}
		}

		req, e := request.CreateFunc(ctx, request.Method, target)
		if e != nil {
			return nil, e
		}

		if e := request.encodeHTTPRequest(ctx, req); e != nil {
			return nil, e
		}

		for _, hook := range c.before {
			ctx = hook.Before(ctx, req)
		}

		resp, e := c.config.HTTPClient.Do(req.WithContext(ctx))
		if e != nil {
			return nil, e
		}
		defer func() { _ = resp.Body.Close() }()

		for _, hook := range c.after {
			ctx = hook.After(ctx, resp)
		}
		return dec(ctx, resp)
	}
}
