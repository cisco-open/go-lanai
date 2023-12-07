package certs_test

import (
	"context"
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"encoding/json"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
)

/*************************
	Test Setup
 *************************/

type ManagerDI struct {
	fx.In
	AppCfg bootstrap.ApplicationConfig
}

func ProvideTestManager(di ManagerDI) (*certs.DefaultManager, error) {
	var props certs.Properties
	if e := di.AppCfg.Bind(&props, "tls"); e != nil {
		return nil, fmt.Errorf("failed to bind TLS properties")
	}
	return certs.NewDefaultManager(func(mgr *certs.DefaultManager) {
		mgr.ConfigLoaderFunc = di.AppCfg.Bind
		mgr.Properties = props
	}), nil
}

type ManagerTestDI struct {
	fx.In
	Manager *certs.DefaultManager
}

func RegisterTestFactories(manager *certs.DefaultManager) {
	manager.MustRegister(&TestSourceFactory{SrcType: certs.SourceFile})
	manager.MustRegister(&TestSourceFactory{SrcType: certs.SourceVault})
	manager.MustRegister(&TestSourceFactory{SrcType: certs.SourceACM})

}

/*************************
	Test
 *************************/

func TestManager(t *testing.T) {
	di := &ManagerTestDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithDI(di),
		apptest.WithFxOptions(
			fx.Provide(ProvideTestManager),
			fx.Invoke(RegisterTestFactories),
		),
		test.GomegaSubTest(SubTestLoadSourceByConfigPath(di), "TestLoadSourceByConfigPath"),
	)
}

/*************************
	SubTest
 *************************/

func SubTestLoadSourceByConfigPath(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var s certs.Source
		s, e = di.Manager.Source(ctx, certs.WithConfigPath("redis.tls.config"))
		g.Expect(e).To(Succeed(), "load source by ConfigPath should not fail")
		g.Expect(s).To(Not(BeNil()), "source by ConfigPath should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource)), "source should be a test source")
		ts := s.(*TestSource)
		g.Expect(ts.Type).To(Equal(certs.SourceVault), "source should be correct type")
		g.Expect(ts.Config).ToNot(BeEmpty(), "source's config should not be empty'")
	}
}

/*************************
	Helpers
 *************************/

type TestSourceFactory struct {
	SrcType certs.SourceType
}

func (f *TestSourceFactory) Type() certs.SourceType {
	return f.SrcType
}

func (f *TestSourceFactory) LoadAndInit(_ context.Context, opts ...certs.SourceOptions) (certs.Source, error) {
	src := certs.SourceConfig{}
	for _, fn := range opts {
		fn(&src)
	}
	var config map[string]interface{}
	if e := json.Unmarshal(src.RawConfig, &config); e != nil {
		return nil, e
	}
	return &TestSource{
		Type:   f.SrcType,
		Config: config,
	}, nil
}

type TestSource struct {
	Type   certs.SourceType
	Config map[string]interface{}
}

func (s *TestSource) TLSConfig(_ context.Context, _ ...certs.TLSOptions) (*tls.Config, error) {
	return nil, fmt.Errorf("dummy source, for test only")
}

func (s *TestSource) Files(_ context.Context) (*certs.CertificateFiles, error) {
	return nil, fmt.Errorf("dummy source, for test only")
}
