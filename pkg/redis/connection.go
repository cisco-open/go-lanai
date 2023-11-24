package redis

import (
	"context"
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//
//	rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

// ConnOptions options for connectivity by manipulating redis.UniversalOptions
type ConnOptions func(opt *redis.UniversalOptions) error

func GetUniversalOptions(p *RedisProperties, opts ...ConnOptions) (*redis.UniversalOptions, error) {
	universal := &redis.UniversalOptions{
		Addrs:              p.Addresses,
		DB:                 p.DB,
		Username:           p.Username,
		Password:           p.Password,
		MaxRetries:         p.MaxRetries,
		MinRetryBackoff:    p.MinRetryBackoff,
		MaxRetryBackoff:    p.MaxRetryBackoff,
		DialTimeout:        p.DialTimeout,
		ReadTimeout:        p.ReadTimeout,
		WriteTimeout:       p.WriteTimeout,
		PoolSize:           p.PoolSize,
		MinIdleConns:       p.MinIdleConns,
		MaxConnAge:         p.MaxConnAge,
		PoolTimeout:        p.PoolTimeout,
		IdleTimeout:        p.IdleTimeout,
		IdleCheckFrequency: p.IdleCheckFrequency,
		// Only cluster clients.
		MaxRedirects:   p.MaxRedirects,
		ReadOnly:       p.ReadOnly,
		RouteByLatency: p.RouteByLatency,
		RouteRandomly:  p.RouteRandomly,

		// The sentinel master name.
		// Only failover clients.
		MasterName:       p.MasterName,
		SentinelPassword: p.SentinelPassword,
	}

	for _, fn := range opts {
		if e := fn(universal); e != nil {
			return nil, e
		}
	}
	return universal, nil
}

func WithDB(dbIndex int) ConnOptions {
	return func(opt *redis.UniversalOptions) error {
		opt.DB = dbIndex
		return nil
	}
}

func WithTLS(ctx context.Context, tc *tlsconfig.ProviderFactory, p tlsconfig.Properties) ConnOptions {
	return func(opt *redis.UniversalOptions) error {
		if tc == nil {
			return fmt.Errorf("TLS auth is enabled for Redis, but TLSProviderFactory is not available")
		}
		t := &tls.Config{} //nolint:gosec // the minVersion is set later on dynamically, so "G402: TLSProperties MinVersion too low." is a false positive
		provider, err := tc.GetProvider(p)
		if err != nil {
			return errors.Wrap(err, "Cannot fetch tls provider")
		}
		t.MinVersion, err = provider.GetMinTlsVersion()
		if err != nil {
			return errors.Wrap(err, "Cannot fetch min tls version from provider")
		}
		t.GetClientCertificate, err = provider.GetClientCertificate(ctx)
		if err != nil {
			return errors.Wrap(err, "Cannot fetch getCertificate func from provider")
		}
		t.RootCAs, err = provider.RootCAs(ctx)
		if err != nil {
			return errors.Wrap(err, "Cannot fetch root CAs from provider")
		}
		opt.TLSConfig = t
		return nil
	}
}

type Client interface {
	redis.UniversalClient
}

type client struct {
	redis.UniversalClient
}
