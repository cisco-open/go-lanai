package embedded

import (
	"context"
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"math/rand"
	"testing"
	"time"
)

var kCtxEmbeddedRedis = struct{}{}

/*******************
	Public
 *******************/

type RedisOptions func(cfg *RedisConfig)
type RedisConfig struct {
	// Port must between 32768 and 65535
	Port int
	// TLS when set, the redis server is run in TLS mode.
	// Note: duo to internal implementation, When running in TLS mode, the Port is ignored
	TLS *tls.Config
}

// Redis start redis at random port (32768-65535) on test package level.
// The actual port get be get using CurrentRedisPort
func Redis(opts ...RedisOptions) suitetest.PackageOptions {
	return suitetest.TestOptions(WithRedis(opts...))
}

// WithRedis start redis at random port (32768-65535) on per test basis
// The actual port get be get using CurrentRedisPort
func WithRedis(opts ...RedisOptions) test.Options {
	//nolint:gosec // Not security related
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	cfg := RedisConfig{
		Port: 0x7fff + r.Intn(0x7fff) + 1,
	}
	for _, fn := range opts {
		fn(&cfg)
	}
	return redisWithConfig(&cfg)
}

func EnableTLS(certs ...func(src *TLSCerts)) RedisOptions {
	tlsCfg, e := ServerTLSWithCerts(certs...)
	if e != nil {
		logger.Warnf(`unable to enable TLS: %v`, e)
	}
	return func(cfg *RedisConfig) {
		cfg.TLS = tlsCfg
	}
}

// RedisWithPort start redis at given port (must between 32768 and 65535) on test package level.
// Deprecated, use Redis(...) to set RedisConfig.Port
func RedisWithPort(port int) suitetest.PackageOptions {
	return Redis(func(cfg *RedisConfig) {
		cfg.Port = port
	})
}

// CurrentRedisPort getter to return embedded redis port. returns -1 if it's not initialized or started
func CurrentRedisPort(ctx context.Context) (port int) {
	port = -1
	srv, ok := ctx.Value(kCtxEmbeddedRedis).(*miniredis.Miniredis)
	if !ok {
		return
	}
	ret := doWithEmbeddedRedis(srv, func(srv *miniredis.Miniredis) interface{} {
		return srv.Server().Addr().Port
	})

	switch v := ret.(type) {
	case int:
		return v
	}
	return
}

/*******************
	Internals
 *******************/

// redisWithConfig start redis based on given RedisConfig
func redisWithConfig(cfg *RedisConfig) test.Options {
	return test.WithOptions(
		test.Setup(func(ctx context.Context, t *testing.T) (context.Context, error) {
			s, e := startEmbeddedRedis(cfg)
			if e != nil {
				return ctx, e
			}
			return context.WithValue(ctx, kCtxEmbeddedRedis, s), nil
		}),
		apptest.WithDynamicProperties(map[string]apptest.PropertyValuerFunc{
			"redis.addrs": func(ctx context.Context) interface{} {
				return fmt.Sprintf("127.0.0.1:%d", CurrentRedisPort(ctx))
			},
		}),
		test.Teardown(func(ctx context.Context, t *testing.T) error {
			if s, ok := ctx.Value(kCtxEmbeddedRedis).(*miniredis.Miniredis); ok {
				stopEmbeddedRedis(s)
			}
			return nil
		}),
	)
}

func startEmbeddedRedis(cfg *RedisConfig) (server *miniredis.Miniredis, err error) {
	switch {
	case cfg.TLS != nil:
		// TLS mode
		server = miniredis.NewMiniRedis()
		err = server.StartTLS(cfg.TLS)
	case cfg.Port <= 0x7fff && cfg.Port != 0:
		err = fmt.Errorf("invalid embedded redis port [%d], should be > 0x7fff", cfg.Port)
	default:
		// Default mode
		server = miniredis.NewMiniRedis()
		addr := fmt.Sprintf("127.0.0.1:%d", cfg.Port)
		err = server.StartAddr(addr)
	}
	if err == nil {
		logger.Infof("Embedded Redis started at %s", server.Addr())
	}
	return
}

func stopEmbeddedRedis(server *miniredis.Miniredis) {
	if server != nil {
		addr := server.Addr()
		server.Close()
		logger.Infof("Embedded Redis stopped at %s", addr)
	}
}

// doWithEmbeddedRedis perform locking on miniredis.Miniredis, bail the operation if server is not started
func doWithEmbeddedRedis(server *miniredis.Miniredis, fn func(srv *miniredis.Miniredis) interface{}) interface{} {
	if server == nil {
		return nil
	}
	server.Lock()
	defer server.Unlock()
	if s := server.Server(); s != nil {
		return fn(server)
	}
	return nil
}
