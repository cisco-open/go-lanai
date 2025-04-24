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
	"github.com/cisco-open/go-lanai/pkg/discovery"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"net/http"
	"net/url"
	"time"
)

type Client interface {
	// Execute send the provided request and parse the response using provided ResponseOptions
	// When using default decoding function:
	// 		- it returns non-nil Response only if the response has 2XX status code
	// 		- it returns non-nil error for 4XX, 5XX status code or any other type of errors
	// 		- the returned error can be casted to *Error
	Execute(ctx context.Context, request *Request, opts ...ResponseOptions) (*Response, error)

	// WithService create a client with specific service with given instance selectors.
	// The returned client is responsible to track service instance changes with help of discovery package,
	// and to perform load-balancing and retrying.
	// The returned client is goroutine-safe and can be reused
	WithService(service string, opts ...SDOptions) (Client, error)

	// WithBaseUrl create a client with specific base URL.
	// The returned client is responsible to perform retrying.
	// The returned client is goroutine-safe and can be reused
	WithBaseUrl(baseUrl string) (Client, error)

	// WithNoTargetResolver expects the request to contain the absolute url
	WithNoTargetResolver() (Client, error)

	// WithConfig create a shallow copy of the client with specified config.
	// Service (with LB) or BaseURL cannot be changed with this method.
	// If any field of provided config is zero value, this value is not applied.
	// The returned client is goroutine-safe and can be reused
	WithConfig(config *ClientConfig) Client
}

// ClientOptions is used for creating Client and its customizers
type ClientOptions func(opt *ClientOption)

// ClientOption carries initial configurations of Clients
type ClientOption struct {
	ClientConfig
	SDClient           discovery.Client
	DefaultSelector    discovery.InstanceMatcher
	DefaultBeforeHooks []BeforeHook
	DefaultAfterHooks  []AfterHook
}

// ClientConfig is used to change Client's config
type ClientConfig struct {
	// HTTPClient underlying http.Client to use
	HTTPClient *http.Client
	// BeforeHooks hooks to use before sending HTTP request
	BeforeHooks []BeforeHook
	// AfterHooks hooks to use before sending HTTP request
	AfterHooks []AfterHook
	// MaxRetries number of retries in case of error. Negative value means no retry.
	// Note: by default, non-2XX response status code error is not retried
	MaxRetries int
	// RetryBackoff time to wait between retries. Negative means retry immediately
	RetryBackoff time.Duration
	// RetryCallback allows fine control when and how to retry.
	// If set, this override MaxRetries and RetryBackoff
	RetryCallback RetryCallback
	// Timeout how long to wait for each execution.
	// Note: this is total duration including RetryBackoff between each attempt, not per-retry timeout.
	Timeout time.Duration
	// Logger used for logging request/response
	Logger log.ContextualLogger
	// Logging configuration of request/response logging
	Logging LoggingConfig
}

type LoggingConfig struct {
	Level           log.LoggingLevel
	DetailsLevel    LogDetailsLevel
	SanitizeHeaders utils.StringSet
	ExcludeHeaders  utils.StringSet
}

type ClientCustomizer interface {
	Customize(opt *ClientOption)
}

type ClientCustomizerFunc func(opt *ClientOption)

func (fn ClientCustomizerFunc) Customize(opt *ClientOption) {
	fn(opt)
}

// Hook is used for intercepting is used for ClientConfig and ClientOptions,
type Hook interface {
	// Before is invoked after the HTTP request is encoded and before the request is sent.
	// The implementing class could also implement order.Ordered interface. Highest order is invoked first
	Before(context.Context, *http.Request) context.Context
	// After is invoked after HTTP response is returned and before the response is decoded.
	// The implementing class could also implement order.Ordered interface. Highest order is invoked first
	After(context.Context, *http.Response) context.Context
}

// BeforeHook is used for ClientConfig and ClientOptions,
// The implementing class could also implement order.Ordered interface. Highest order is invoked first
type BeforeHook interface {
	// Before is invoked after the HTTP request is encoded and before the request is sent.
	Before(context.Context, *http.Request) context.Context
}

// ConfigurableBeforeHook is an additional interface that BeforeHook can implement
type ConfigurableBeforeHook interface {
	WithConfig(cfg *ClientConfig) BeforeHook
}

// AfterHook is used for ClientConfig and ClientOptions,
// The implementing class could also implement order.Ordered interface. Highest order is invoked first
type AfterHook interface {
	// After is invoked after HTTP response is returned and before the response is decoded.
	After(context.Context, *http.Response) context.Context
}

// ConfigurableAfterHook is an additional interface that AfterHook can implement
type ConfigurableAfterHook interface {
	WithConfig(cfg *ClientConfig) AfterHook
}

type TargetResolver interface {
	Resolve(ctx context.Context, req *Request) (*url.URL, error)
}

type TargetResolverFunc func(ctx context.Context, req *Request) (*url.URL, error)

func (fn TargetResolverFunc) Resolve(ctx context.Context, req *Request) (*url.URL, error) {
	return fn(ctx, req)
}

/************************
	Common Impl.
 ************************/

func DefaultConfig() *ClientConfig {
	return &ClientConfig{
		HTTPClient:  http.DefaultClient,
		BeforeHooks: []BeforeHook{},
		AfterHooks:  []AfterHook{},
		MaxRetries:  3,
		Timeout:     1 * time.Minute,
		Logger:      logger,
		Logging: LoggingConfig{
			DetailsLevel:    LogDetailsLevelHeaders,
			SanitizeHeaders: utils.NewStringSet(HeaderAuthorization),
			ExcludeHeaders:  utils.NewStringSet(),
		},
	}
}

// defaultServiceConfig add necessary configs/hooks for internal load balanced service
func defaultServiceConfig() *ClientConfig {
	return &ClientConfig{
		BeforeHooks: []BeforeHook{HookTokenPassthrough()},
	}
}

// defaultExtHostConfig add necessary configs/hooks for external hosts
func defaultExtHostConfig() *ClientConfig {
	return &ClientConfig{}
}

func mergeConfig(dst *ClientConfig, src *ClientConfig) {
	if dst.HTTPClient == nil {
		dst.HTTPClient = src.HTTPClient
	}

	if dst.Logger == nil {
		dst.Logger = src.Logger
	}

	if dst.Timeout <= 0 {
		dst.Timeout = src.Timeout
	}

	if dst.BeforeHooks == nil {
		dst.BeforeHooks = src.BeforeHooks
	}

	if dst.AfterHooks == nil {
		dst.AfterHooks = src.AfterHooks
	}

	switch {
	case dst.MaxRetries < 0:
		dst.MaxRetries = 0
	case dst.MaxRetries == 0:
		dst.MaxRetries = src.MaxRetries
	}

	switch {
	case dst.RetryBackoff < 0:
		dst.RetryBackoff = 0
	case dst.RetryBackoff == 0:
		dst.RetryBackoff = src.RetryBackoff
	}

	if dst.RetryCallback == nil {
		dst.RetryCallback = src.RetryCallback
	}

	if dst.Logging.SanitizeHeaders == nil {
		dst.Logging.SanitizeHeaders = src.Logging.SanitizeHeaders
	}

	if dst.Logging.ExcludeHeaders == nil {
		dst.Logging.ExcludeHeaders = src.Logging.ExcludeHeaders
	}

	if dst.Logging.DetailsLevel == LogDetailsLevelUnknown {
		dst.Logging.DetailsLevel = src.Logging.DetailsLevel
	}

	if dst.Logging.Level == log.LevelOff {
		dst.Logging.Level = src.Logging.Level
	}
}
