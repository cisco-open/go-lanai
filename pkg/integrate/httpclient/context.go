package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	httptransport "github.com/go-kit/kit/transport/http"
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
	BeforeHooks []BeforeHook
	AfterHooks  []AfterHook
	MaxRetries  int // negative value means no retry
	Timeout     time.Duration
	Logger      log.ContextualLogger
	Verbose     bool
}

type ClientCustomizer interface {
	Customize(opt *ClientOption)
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
		BeforeHooks: []BeforeHook{},
		AfterHooks:  []AfterHook{},
		MaxRetries:  3,
		Timeout:     1 * time.Minute,
		Logger:      logger,
		Verbose:     false,
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
