package certs_test

import (
	"context"
	"crypto/tls"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	certsource "cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs/source"
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

func ProvideTestManager(di ManagerDI) (*certs.DefaultManager, certs.Properties, error) {
	var props certs.Properties
	if e := di.AppCfg.Bind(&props, "certificates"); e != nil {
		return nil, props, fmt.Errorf("failed to bind TLS properties")
	}
	manager := certs.NewDefaultManager(func(mgr *certs.DefaultManager) {
		mgr.ConfigLoaderFunc = di.AppCfg.Bind
		mgr.Properties = props
	})
	manager.MustRegister(NewFileSourceFactory(props))
	manager.MustRegister(NewVaultSourceFactory(props))
	return manager, props, nil
}

type ManagerTestDI struct {
	fx.In
	Manager *certs.DefaultManager
	AppCfg  bootstrap.ApplicationConfig
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
		),
		test.GomegaSubTest(SubTestLoadSourceByPreset(di), "TestLoadSourceByPreset"),
		test.GomegaSubTest(SubTestLoadSourceByConfigPath(di), "TestLoadSourceByConfigPath"),
		test.GomegaSubTest(SubTestLoadSourceByRawConfig(di), "TestLoadSourceByRawConfig"),
		test.GomegaSubTest(SubTestLoadSourceByCompatibleConfig(di), "TestLoadSourceByCompatibleConfig"),
		test.GomegaSubTest(SubTestLoadSourceBySourceProperties(di), "TestLoadSourceBySourceProperties"),
		test.GomegaSubTest(SubTestLoadSourceWithInvalidOptions(di), "TestLoadSourceWithInvalidOptions"),
		test.GomegaSubTest(SubTestCloseSources(di), "TestCloseSources"),
	)
}

/*************************
	SubTest
 *************************/

func SubTestLoadSourceByPreset(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var s certs.Source
		s, e = di.Manager.Source(ctx, certs.WithPreset("redis-file"))
		g.Expect(e).To(Succeed(), "load source by Preset should not fail")
		g.Expect(s).To(Not(BeNil()), "source by Preset should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "preset")
	}
}

func SubTestLoadSourceByConfigPath(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var s certs.Source
		s, e = di.Manager.Source(ctx, certs.WithConfigPath("data.cockroach.tls.certs"))
		g.Expect(e).To(Succeed(), "load source by ConfigPath should not fail")
		g.Expect(s).To(Not(BeNil()), "source by ConfigPath should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")
	}
}

func SubTestLoadSourceByRawConfig(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var raw json.RawMessage
		var s certs.Source
		e = di.AppCfg.Bind(&raw, "data.cockroach.tls.certs")
		g.Expect(e).To(Succeed(), "parse raw config should not fail")

		// as json.RawMessage
		s, e = di.Manager.Source(ctx, certs.WithRawConfig(raw))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")

		// as []byte
		s, e = di.Manager.Source(ctx, certs.WithRawConfig([]byte(raw)))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")

		// as string
		s, e = di.Manager.Source(ctx, certs.WithRawConfig(string(raw)))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")
	}
}

func SubTestLoadSourceByCompatibleConfig(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var compatible = struct {
			TestFileSourceProperties
		}{}
		var e error
		var s certs.Source
		e = di.AppCfg.Bind(&compatible, "data.cockroach.tls.certs")
		g.Expect(e).To(Succeed(), "parse raw config should not fail")

		s, e = di.Manager.Source(ctx, certs.WithType(certs.SourceFile, compatible))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")
	}
}

func SubTestLoadSourceBySourceProperties(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var s certs.Source
		// Kafka
		var kafkaProps TestKafkaProperties
		e = di.AppCfg.Bind(&kafkaProps, "kafka")
		g.Expect(e).To(Succeed(), "load 'kafka' properties should not fail")

		s, e = di.Manager.Source(ctx, certs.WithSourceProperties(&kafkaProps.TLS.Certs))
		g.Expect(e).To(Succeed(), "load source by kafka's SourceProperties should not fail")
		g.Expect(s).To(Not(BeNil()), "source by kafka's SourceProperties should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "source-default")

		// Redis
		var redisProps TestRedisProperties
		e = di.AppCfg.Bind(&redisProps, "redis")
		g.Expect(e).To(Succeed(), "load 'redis' properties should not fail")

		s, e = di.Manager.Source(ctx, certs.WithSourceProperties(&redisProps.TLS.Certs))
		g.Expect(e).To(Succeed(), "load source by redis' SourceProperties should not fail")
		g.Expect(s).To(Not(BeNil()), "source by redis' SourceProperties should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "preset")

		// Data
		var dataProps TestDataProperties
		e = di.AppCfg.Bind(&dataProps, "data.cockroach")
		g.Expect(e).To(Succeed(), "load `data.cockroach` properties should not fail")

		s, e = di.Manager.Source(ctx, certs.WithSourceProperties(&dataProps.TLS.Certs))
		g.Expect(e).To(Succeed(), "load source by data's SourceProperties should not fail")
		g.Expect(s).To(Not(BeNil()), "source by data's SourceProperties should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestSource[TestFileSourceProperties])), "source's type should be correct")
		AssertFileSourceProperties(g, s.(*TestSource[TestFileSourceProperties]).Config, "adhoc")
	}
}

func SubTestLoadSourceWithInvalidOptions(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		// no options
		_, e = di.Manager.Source(ctx)
		g.Expect(e).To(HaveOccurred(), "load source without any option should fail")

		// incompatible raw config
		_, e = di.Manager.Source(ctx, certs.WithRawConfig(nomarshaler{}))
		g.Expect(e).To(HaveOccurred(), "load source by incompatible raw config should fail")

		// non-existing config path
		_, e = di.Manager.Source(ctx, certs.WithConfigPath("does.not.exist"))
		g.Expect(e).To(HaveOccurred(), "load source by non-existing config path should fail")

		// conflicting options
		_, e = di.Manager.Source(ctx, certs.WithPreset("redis-file"), certs.WithRawConfig(`{"type":"file"}`))
		g.Expect(e).To(HaveOccurred(), "load source with conflicting options should fail")

		// no type
		_, e = di.Manager.Source(ctx, certs.WithRawConfig(`{"no-type":"oops"}`))
		g.Expect(e).To(HaveOccurred(), "load source without should fail")

		// wrong type
		_, e = di.Manager.Source(ctx, certs.WithRawConfig(`{"type":"doesn't exists'"}`))
		g.Expect(e).To(HaveOccurred(), "load source with non-existing source type should fail")
	}
}

func SubTestCloseSources(di *ManagerTestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		vaultSrcProps := TestVaultSourceProperties{Path: "/pki"}
		var e error
		var s certs.Source
		var srcs []*TestClosableSource[TestVaultSourceProperties]
		// 1
		s, e = di.Manager.Source(ctx, certs.WithType(certs.SourceVault, vaultSrcProps))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestClosableSource[TestVaultSourceProperties])), "source's type should be correct")
		srcs = append(srcs, s.(*TestClosableSource[TestVaultSourceProperties]))
		// 2
		s, e = di.Manager.Source(ctx, certs.WithType(certs.SourceVault, vaultSrcProps))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestClosableSource[TestVaultSourceProperties])), "source's type should be correct")
		srcs = append(srcs, s.(*TestClosableSource[TestVaultSourceProperties]))
		// 3
		s, e = di.Manager.Source(ctx, certs.WithType(certs.SourceVault, vaultSrcProps))
		g.Expect(e).To(Succeed(), "load source by RawConfig should not fail")
		g.Expect(s).To(Not(BeNil()), "source by RawConfig should not be nil")
		g.Expect(s).To(BeAssignableToTypeOf(new(TestClosableSource[TestVaultSourceProperties])), "source's type should be correct")
		srcs = append(srcs, s.(*TestClosableSource[TestVaultSourceProperties]))

		// Test Closer
		e = di.Manager.Close()
		g.Expect(e).To(Succeed(), "Closing manager should not fail")
		for _, closeable := range srcs {
			g.Expect(closeable.Closed).To(BeTrue(), "source should be closed")
		}
	}
}

/*************************
	Helpers
 *************************/

func AssertFileSourceProperties(g *gomega.WithT, props TestFileSourceProperties, expectedFilename string) {
	g.Expect(props).ToNot(BeZero(), "source properties should not be empty")
	g.Expect(props.CACertFile).To(HaveSuffix("ca.crt"), "source's CA path should be correct")
	g.Expect(props.CertFile).To(HaveSuffix(expectedFilename+".crt"), "source's Cert path should be correct")
	g.Expect(props.KeyFile).To(HaveSuffix(expectedFilename+".key"), "source's Key path should be correct")
}

func NewFileSourceFactory(props certs.Properties) *certsource.GenericFactory[TestFileSourceProperties] {
	defaults := props.Sources[certs.SourceFile]
	factory, e := certsource.NewFactory(certs.SourceFile, defaults, NewTestFileSource)
	if e != nil {
		return nil
	}
	return factory
}

func NewVaultSourceFactory(props certs.Properties) *certsource.GenericFactory[TestVaultSourceProperties] {
	defaults := props.Sources[certs.SourceVault]
	factory, e := certsource.NewFactory(certs.SourceVault, defaults, NewTestVaultSource)
	if e != nil {
		return nil
	}
	return factory
}

func NewTestFileSource(props TestFileSourceProperties) certs.Source {
	return &TestSource[TestFileSourceProperties]{
		Config: props,
	}
}

func NewTestVaultSource(props TestVaultSourceProperties) certs.Source {
	return &TestClosableSource[TestVaultSourceProperties]{
		TestSource: TestSource[TestVaultSourceProperties]{
			Config: props,
		},
	}
}

type TestSource[T any] struct {
	Config T
}

func (s *TestSource[T]) TLSConfig(_ context.Context, _ ...certs.TLSOptions) (*tls.Config, error) {
	return nil, fmt.Errorf("dummy source, for test only")
}

func (s *TestSource[T]) Files(_ context.Context) (*certs.CertificateFiles, error) {
	return nil, fmt.Errorf("dummy source, for test only")
}

type TestClosableSource[T any] struct {
	TestSource[T]
	Closed bool
}

func (s *TestClosableSource[T]) Close() error {
	if s.Closed {
		return fmt.Errorf("already closed")
	}
	s.Closed = true
	return nil
}

type nomarshaler struct{}

func (nomarshaler) MarshalJSON() ([]byte, error) {
	return nil, fmt.Errorf("JSON no love")
}
