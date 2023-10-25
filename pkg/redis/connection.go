package redis

import (
	"context"
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//
//	rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

func GetUniversalOptions(ctx context.Context, p *RedisProperties, tc *tlsconfig.ProviderFactory) (*redis.UniversalOptions, error) {
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
	if p.Tls.Enable {
		t := &tls.Config{} //nolint:gosec // the minVersion is set later on dynamically, so "G402: TLS MinVersion too low." is a false positive
		provider, err := tc.GetProvider(p.Tls.Config)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot fetch tls provider")
		}
		t.MinVersion, err = provider.GetMinTlsVersion()
		if err != nil {
			return nil, errors.Wrap(err, "Cannot fetch min tls version from provider")
		}
		t.GetClientCertificate, err = provider.GetClientCertificate(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot fetch getCertificate func from provider")
		}
		t.RootCAs, err = provider.RootCAs(ctx)
		if err != nil {
			return nil, errors.Wrap(err, "Cannot fetch root CAs from provider")
		}
		universal.TLSConfig = t
	}

	return universal, nil
}

type Client interface {
	redis.UniversalClient
}

type client struct {
	redis.UniversalClient
}
