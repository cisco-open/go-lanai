package redis

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

type ClientOptions func(opt *ClientOption)

type ClientOption struct {
	DbIndex int
}

type OptionsAwareHook interface {
	redis.Hook
	WithClientOption(*redis.UniversalOptions) redis.Hook
}

type ClientFactory interface {
	// New returns a newly created Client
	New(ctx context.Context, opts ...ClientOptions) (Client, error)

	// AddHooks add hooks to all Client already created and any future Client created via this interface
	// If the given hook also implments OptionsAwareHook, the method will be used to derive a hook instance and added to
	// coresponding client
	AddHooks(ctx context.Context, hooks ...redis.Hook)
}

// clientFactory implements ClientFactory
type clientRecord struct {
	client  Client
	options *redis.UniversalOptions
}

type clientFactory struct {
	properties   RedisProperties
	hooks        []redis.Hook
	clients      map[ClientOption]clientRecord
	certsManager tlsconfig.Manager
}

type FactoryOptions func(opt *FactoryOption)
type FactoryOption struct {
	Properties      RedisProperties
	TLSCertsManager tlsconfig.Manager
}

func NewClientFactory(opts...FactoryOptions) ClientFactory {
	opt := FactoryOption{}
	for _, fn := range opts {
		fn(&opt)
	}
	return &clientFactory{
		properties:   opt.Properties,
		hooks:        []redis.Hook{},
		clients:      map[ClientOption]clientRecord{},
		certsManager: opt.TLSCertsManager,
	}
}

func (f *clientFactory) New(ctx context.Context, opts ...ClientOptions) (Client, error) {
	opt := ClientOption{}
	for _, f := range opts {
		f(&opt)
	}

	// Some validations
	if opt.DbIndex < 0 || opt.DbIndex >= 16 {
		return nil, fmt.Errorf("invalid Redis DB index [%d]: must be between 0 and 16", opt.DbIndex)
	}

	if existing, ok := f.clients[opt]; ok {
		return existing.client, nil
	}

	// TODO review this part for more reasonable API design
	connOpts := []ConnOptions{withDB(opt.DbIndex)}
	if f.properties.TLS.Enabled {
		connOpts = append(connOpts, withTLS(ctx, f.certsManager, &f.properties.TLS.Config))
	}

	// prepare options
	options, e := GetUniversalOptions(&f.properties, connOpts...)
	if e != nil {
		return nil, errors.Wrap(e, "Invalid redis configuration")
	}

	c := client{
		UniversalClient: redis.NewUniversalClient(options),
	}

	// apply hooks
	for _, hook := range f.hooks {
		h := hook
		if aware, ok := hook.(OptionsAwareHook); ok {
			h = aware.WithClientOption(options)
		}
		c.AddHook(h)
	}

	// record the client
	f.clients[opt] = clientRecord{
		client:  c,
		options: options,
	}

	logger.WithContext(ctx).Infof("Redis client created with DB index %d", options.DB)
	return &c, nil
}

func (f *clientFactory) AddHooks(ctx context.Context, hooks ...redis.Hook) {
	f.hooks = append(f.hooks, hooks...)
	// add to existing clients
	for _, hook := range hooks {
		for _, record := range f.clients {
			h := hook
			if aware, ok := hook.(OptionsAwareHook); ok {
				h = aware.WithClientOption(record.options)
			}
			record.client.AddHook(h)
		}
	}
	logger.WithContext(ctx).Debugf("Added redis hooks: %v", hooks)
}
