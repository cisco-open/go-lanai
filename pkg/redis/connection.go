package redis

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
)

// KeepTTL is an option for Set command to keep key's existing TTL.
// For example:
//
//    rdb.Set(ctx, key, value, redis.KeepTTL)
const KeepTTL = redis.KeepTTL

func GetUniversalOptions(p *RedisProperties) (*redis.UniversalOptions, error) {
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

	if p.RootCertificates != "" {
		file, err := os.Open(p.RootCertificates)

		if err != nil {
			return nil, errors.Wrap(err, "Cannot open root certificates file: "+p.RootCertificates)
		}

		data, err := ioutil.ReadAll(file)

		if err != nil {
			return nil, errors.Wrap(err, "Cannot read root certificates file: "+p.RootCertificates)
		}

		root := x509.NewCertPool()
		ok := root.AppendCertsFromPEM(data)

		if !ok {
			return nil, errors.New("Cannot parse the certificate file content")
		}

		t := &tls.Config{
			RootCAs: root,
			MinVersion: tls.VersionTLS12,
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
