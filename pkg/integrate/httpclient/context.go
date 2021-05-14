package httpclient

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"github.com/go-kit/kit/endpoint"
	"io"
)

type EndpointFactory func(inst *discovery.Instance) (endpoint.Endpoint, io.Closer, error)
