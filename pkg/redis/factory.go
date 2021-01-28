package redis

import (
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type ClientOptions func(opt *ClientOption)

type ClientOption struct {
	DbIndex int
}

type ClientFactory interface {
	New(opts ...ClientOptions) (Client, error)
}

// clientFactory implements ClientFactory
type clientFactory struct {
	properties ConnectionProperties
}

func NewClientFactory(p ConnectionProperties) ClientFactory {
	return &clientFactory{
		properties: p,
	}
}

func (f *clientFactory) New(opts ...ClientOptions) (Client, error) {
	opt := ClientOption{}
	for _, f := range opts {
		f(&opt)
	}

	// Some validations
	if opt.DbIndex < 0 || opt.DbIndex >= 16 {
		return nil, fmt.Errorf("invalid Redis DB index [%d]: must be between 0 and 16", opt.DbIndex)
	}

	// prepare options
	options, e := GetUniversalOptions(&f.properties)
	if e != nil {
		return nil, errors.Wrap(e, "Invalid redis configuration")
	}

	// customize
	options.DB = opt.DbIndex

	c := client {
		UniversalClient: redis.NewUniversalClient(options),
	}

	return &c, nil
}
