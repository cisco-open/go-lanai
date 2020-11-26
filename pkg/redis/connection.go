package redis

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"time"
)

const (
	ConfigRootRedisConnection = "redis"
)

type ConnectionProperties struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addrs []string `json:"addrs"`

	// Database to be selected after connecting to the server.
	// Only single-node and failover clients.
	DB int `json:"db"`

	// Common options.
	Username string `json:"username"`
	Password string `json:"password"`

	MaxRetries      int           `json:"max-retries"`
	MinRetryBackoff time.Duration `json:"min-retry-backoff"`
	MaxRetryBackoff time.Duration `json:"max-retry-backoff"`

	DialTimeout  time.Duration `json:"dial-timeout"`
	ReadTimeout  time.Duration `json:"read-timeout"`
	WriteTimeout time.Duration `json:"write-timeout"`

	PoolSize           int           `json:"pool-size"`
	MinIdleConns       int           `json:"min-idle-conns"`
	MaxConnAge         time.Duration `json:"max-conn-age"`
	PoolTimeout        time.Duration `json:"pool-timeout"`
	IdleTimeout        time.Duration `json:"idle-timeout"`
	IdleCheckFrequency time.Duration `json:"idle-check-frequency"`

	//path to root certificates files
	RootCertificates string `json:"root-certificates"`

	// Only cluster clients.

	MaxRedirects   int  `json:"max-redirects"`
	ReadOnly       bool `json:"read-only"`
	RouteByLatency bool `json:"route-by-latency"`
	RouteRandomly  bool `json:"route-randomly"`

	// The sentinel master name.
	// Only failover clients.
	MasterName       string `json:"master-name"`
	SentinelPassword string `json:"sentinel-password"`
}

func GetUniversalOptions(p *ConnectionProperties) (*redis.UniversalOptions, error) {
	universal := &redis.UniversalOptions{
		Addrs:              p.Addrs,
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

		t := &tls.Config{RootCAs: root}
		universal.TLSConfig = t
	}
	return universal, nil
}

type Connection struct {
	redis.UniversalClient
}
