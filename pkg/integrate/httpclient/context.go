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
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	httptransport "github.com/go-kit/kit/transport/http"
	"net/http"
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
	WithService(service string, selectors ...discovery.InstanceMatcher) (Client, error)

	// WithBaseUrl create a client with specific base URL.
	// The returned client is responsible to perform retrying.
	// The returned client is goroutine-safe and can be reused
	WithBaseUrl(baseUrl string) (Client, error)

	// WithConfig create a shallow copy of the client with specified config.
	// Service (with LB) or BaseURL cannot be changed with this method.
	// If non-primitive field of provided config is zero value, this value is not applied.
	// The returned client is goroutine-safe and can be reused
	WithConfig(config *ClientConfig) Client
}

// ClientOptions is used for creating Client and its customizers
type ClientOptions func(opt *ClientOption)

// ClientOption carries initial configurations of Clients
type ClientOption struct {
	ClientConfig
	DefaultSelector    discovery.InstanceMatcher
	DefaultBeforeHooks []BeforeHook
	DefaultAfterHooks  []AfterHook
}

// ClientConfig is used to change Client's config
type ClientConfig struct {
	HTTPClient  *http.Client // underlying http.Client to use
	BeforeHooks []BeforeHook
	AfterHooks  []AfterHook
	MaxRetries  int // negative value means no retry
	Timeout     time.Duration
	Logger      log.ContextualLogger
	Logging     LoggingConfig
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

// BeforeHook is used for ClientConfig and ClientOptions, the RequestFunc is invoked before request is sent
// implementing class could also implement order.Ordered interface. Highest order is invoked first
type BeforeHook interface {
	RequestFunc() httptransport.RequestFunc
}

// AfterHook is used for ClientConfig and ClientOptions, the ResponseFunc is invoked after response is returned
// implementing class could also implement order.Ordered interface. Highest order is invoked first
type AfterHook interface {
	ResponseFunc() httptransport.ClientResponseFunc
}

// ConfigurableBeforeHook is an additional interface that BeforeHook can implement
type ConfigurableBeforeHook interface {
	WithConfig(cfg *ClientConfig) BeforeHook
}

// ConfigurableAfterHook is an additional interface that AfterHook can implement
type ConfigurableAfterHook interface {
	WithConfig(cfg *ClientConfig) AfterHook
}

// EndpointFactory takes a instance descriptor and create endpoint.Endpoint
// Supported instance type could be :
//		- *discovery.Instance
//		- *url.URL as base url
type EndpointFactory func(instDesp interface{}) (endpoint.Endpoint, error)

type Endpointer interface {
	sd.Endpointer
	WithConfig(config *EndpointerConfig) Endpointer
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
