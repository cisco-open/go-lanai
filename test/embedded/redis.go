package embedded

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"fmt"
	"github.com/alicebob/miniredis/v2"
	"math/rand"
	"time"
)

var embeddedRedis *miniredis.Miniredis

// Redis start redis at random port (32768-65535) on test package level.
// The actual port get be get using CurrentRedisPort
// See RedisWithPort for more details
func Redis() suitetest.PackageOptions {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	port := 0x7fff + r.Intn(0x7fff) + 1
	return RedisWithPort(port)
}

// RedisWithPort start redis at given port (must between 32768 and 65535) on test package level.
func RedisWithPort(port int) suitetest.PackageOptions {
	return suitetest.WithOptions(
		suitetest.SetupWithOrder(orderEmbeddedRedis, startEmbeddedRedisFunc(port, &embeddedRedis)),
		suitetest.TestOptions(
			apptest.WithDynamicProperties(map[string]apptest.PropertyValuerFunc{
				"redis.addrs": func(ctx context.Context) interface{} {
					return fmt.Sprintf("127.0.0.1:%d", CurrentRedisPort())
				},
			}),
		),
		suitetest.TeardownWithOrder(orderEmbeddedRedis, stopEmbeddedRedisFunc(&embeddedRedis)),
	)
}

// CurrentRedisPort getter to return embedded redis port. returns -1 if it's not initialized or started
func CurrentRedisPort() int {
	ret := doWithEmbeddedRedis(func(srv *miniredis.Miniredis) interface{} {
		return srv.Server().Addr().Port
	})
	switch v := ret.(type) {
	case int:
		return v
	default:
		return -1
	}
}

func startEmbeddedRedisFunc(port int, serverPtr **miniredis.Miniredis) suitetest.SetupFunc {
	return func() error {
		if port <= 0x7fff {
			return fmt.Errorf("invalid embedded redis port [%d], should be > 0x7fff", port)
		}

		addr := fmt.Sprintf("127.0.0.1:%d", port)
		s := miniredis.NewMiniRedis()
		if e := s.StartAddr(addr); e != nil {
			return e
		}
		*serverPtr = s
		logger.Infof("Embedded Redis started at %s", addr)
		return nil
	}
}

func stopEmbeddedRedisFunc(server **miniredis.Miniredis) suitetest.TeardownFunc {
	return func() error {
		if server != nil && *server != nil {
			addr := (*server).Addr()
			(*server).Close()
			logger.Infof("Embedded Redis stopped at %s", addr)
		}
		return nil
	}
}

// doWithEmbeddedRedis perform locking on miniredis.Miniredis, bail the operation if server is not started
func doWithEmbeddedRedis(fn func(srv *miniredis.Miniredis) interface{}) interface{} {
	if embeddedRedis == nil {
		return nil
	}
	embeddedRedis.Lock()
	defer embeddedRedis.Unlock()
	if s := embeddedRedis.Server(); s != nil {
		return fn(embeddedRedis)
	}
	return nil
}
