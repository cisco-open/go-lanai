package tlsconfig_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tlsconfig"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Setup
 *************************/

type ManagerTestDI struct {
	fx.In
	AppConfig bootstrap.ApplicationConfig
	Manager   *tlsconfig.DefaultManager
}

func RegisterTestFactories(manager *tlsconfig.DefaultManager) {
	manager.MustRegister(&TestProviderFactory{T: tlsconfig.SourceFile})
	manager.MustRegister(&TestProviderFactory{T: tlsconfig.SourceVault})
	manager.MustRegister(&TestProviderFactory{T: tlsconfig.SourceACM})

}

/*************************
	Test
 *************************/

func TestManager(t *testing.T) {
	di := &ManagerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		//apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(tlsconfig.NewDefaultManager),
			fx.Invoke(RegisterTestFactories),
		),
		test.GomegaSubTest(SubTestLoadProperties(di), "TestFileProvider"),
		test.GomegaSubTest(SubTestLoadProviderByConfigPath(di), "TestLoadProviderByConfigPath"),
	)
}

/*************************
	SubTest
 *************************/

func SubTestLoadProperties(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		cfg := tlsconfig.SourceConfig{}
		e = di.AppConfig.Bind(&cfg, "tls.sources.vault")
		g.Expect(e).To(Succeed(), "bind vault source should not fail")

		cfg = tlsconfig.SourceConfig{}
		e = di.AppConfig.Bind(&cfg, "tls.sources.file")
		g.Expect(e).To(Succeed(), "bind file source should not fail")
	}
}

func SubTestLoadProviderByConfigPath(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var p tlsconfig.Provider
		p, e = di.Manager.Provider(ctx, func(opt *tlsconfig.Option) {
			opt.ConfigPath = "redis.tls.config"
		})
		g.Expect(e).To(Succeed(), "get provider by ConfigPath should not fail")
		g.Expect(p).To(Not(BeNil()), "provider by ConfigPath should not be nil")
	}
}

/*************************
	Helpers
 *************************/

type TestProviderFactory struct {
	T tlsconfig.SourceType
}

func (f *TestProviderFactory) Type() tlsconfig.SourceType {
	return f.T
}

func (f *TestProviderFactory) LoadAndInit(_ context.Context, opts ...tlsconfig.SourceOptions) (tlsconfig.Provider, error) {
	src := tlsconfig.SourceConfig{}
	for _, fn := range opts {
		fn(&src)
	}
	var props tlsconfig.Properties
	if e := json.Unmarshal(src.RawConfig, &props); e != nil {
		return nil, e
	}
	// TODO
	return nil, errors.New(fmt.Sprintf("%s based tls config provider is not supported", props.Type))
}

