package httpclient

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
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
	// Service (with LB) or BaseURL cannot be changed with this method
	// The returned client is goroutine-safe and can be reused
	WithConfig(config *ClientConfig) Client
}

type EndpointFactory func(inst *discovery.Instance) (endpoint.Endpoint, error)

type Endpointer interface {
	sd.Endpointer
	WithConfig(config *EndpointerConfig) Endpointer
}


