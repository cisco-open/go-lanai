package redis_examples

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/embedded"
	"fmt"
	goredis "github.com/go-redis/redis/v8"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Examples
 *************************/

// TestMain is the alternative place we could kick off embedded redis at the package level
//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		embedded.Redis(),
//	)
//}

func TestRedisWithoutApp(t *testing.T) {
	test.RunTest(context.Background(), t,
		// Kick off embedded redis at test level
		embedded.WithRedis(),
		test.GomegaSubTest(SubTestExampleWithoutApp(), "SubTestWithoutApp"),
	)
}

type redisDI struct {
	fx.In
	DefaultClient redis.Client        `optional:"true"`
	ClientFactory redis.ClientFactory `optional:"true"`
}

func TestRedisWithApp(t *testing.T) {
	di := &redisDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		// Kick off embedded redis at test level
		embedded.WithRedis(),
		apptest.WithModules(redis.Module),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestExampleWithApp(di), "SubTestWithApp"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestExampleWithoutApp() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// create an simple client
		universal := &goredis.UniversalOptions{}
		opts := universal.Simple()
		opts.Addr = fmt.Sprintf("127.0.0.1:%d", embedded.CurrentRedisPort(ctx))
		client := goredis.NewClient(opts)
		defer func() { _ = client.Close() }()

		// ping
		cmd := client.Ping(ctx)
		g.Expect(cmd).To(Not(BeNil()), "redis ping shouldn't return nil")
		g.Expect(cmd.Err()).To(Succeed(), "redis ping shouldn't return error")
	}
}

func SubTestExampleWithApp(di *redisDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(di.DefaultClient).To(Not(BeNil()), "injected default redis.Client should not be nil")
		g.Expect(di.ClientFactory).To(Not(BeNil()), "injected redis.ClientFactory should not be nil")

		// use factory
		client, e := di.ClientFactory.New(ctx, func(opt *redis.ClientOption) {
			opt.DbIndex = 5
		})
		g.Expect(e).To(Succeed(), "injected client factory shouldn't return error")

		// ping
		cmd := client.Ping(ctx)
		g.Expect(cmd).To(Not(BeNil()), "redis ping shouldn't return nil")
		g.Expect(cmd.Err()).To(Succeed(), "redis ping shouldn't return error")
	}
}
