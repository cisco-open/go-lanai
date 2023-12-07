package redis

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/pkg/errors"
	"time"
)

const (
	ConfigRootRedisConnection = "redis"
	DefaultDbIndex            = 0
)

type RedisProperties struct {
	// Either a single address or a seed list of host:port addresses
	// of cluster/sentinel nodes.
	Addresses utils.CommaSeparatedSlice `json:"addrs"`

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
	//RootCertificates string `json:"root-certificates"`
	// TLS Properties for Redis
	TLS TLSProperties `json:"tls"`
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

type TLSProperties struct {
	Enabled bool                   `json:"enabled"`
	Config  certs.SourceProperties `json:"config"`
}

func BindRedisProperties(ctx *bootstrap.ApplicationContext) RedisProperties {
	props := RedisProperties{}
	if err := ctx.Config().Bind(&props, ConfigRootRedisConnection); err != nil {
		panic(errors.Wrap(err, "failed to bind redis.RedisProperties"))
	}
	return props
}
