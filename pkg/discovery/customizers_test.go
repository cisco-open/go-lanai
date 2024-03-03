package discovery

import (
	"context"
	"embed"
	"github.com/cisco-open/go-lanai/pkg/bootstrap"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"testing"
	"time"
)

//go:embed testdata/bootstrap-test.yml
var TestBootstrapFS embed.FS

type testDI struct {
	fx.In
	AppCtx *bootstrap.ApplicationContext
}

func TestCustomizers(t *testing.T) {
	var di testDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithBootstrapConfigFS(TestBootstrapFS),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestBuildInfo(), "BuildInfo"),
		test.GomegaSubTest(SubTestPropertiesBased(&di), "PropertiesBased"),
	)
}

func SubTestBuildInfo() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		now := time.Now().Format(utils.ISO8601Seconds)
		bootstrap.BuildVersion = "0.0.1-255"
		bootstrap.BuildTime = now

		customizer := NewBuildInfoCustomizer()
		reg := &TestServiceRegistration{}
		customizer.Customize(ctx, reg)
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagBuildVersion, "0.0.1-255"))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagBuildDateTime, now))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagBuildNumber, "255"))
		g.Expect(reg.Tags()).To(ContainElements(
			KVTag(TagBuildVersion, "0.0.1-255"),
			KVTag(TagBuildDateTime, now),
			KVTag(TagBuildNumber, "255"),
		))
	}
}

func SubTestPropertiesBased(di *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		customizer := NewPropertiesBasedCustomizer(di.AppCtx, map[string]string{
			`name`:         `application.name`,
			`context-path`: `server.context-path`,
			`custom`:       `test.key`,
		})
		reg := &TestServiceRegistration{}
		customizer.Customize(ctx, reg)
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagInstanceUUID, Not(BeZero())))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagServiceName, "test-app"))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(TagContextPath, "/test-api"))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(`context-path`, "/test-api"))
		g.Expect(reg.Meta()).To(HaveKeyWithValue(`custom`, "value"))
		g.Expect(reg.Tags()).To(ContainElements(
			MatchRegexp(`componentAttributes=[^=]+`),
			MatchRegexp(`instanceUuid=[a-f0-9\-]+`),
			KVTag(TagServiceName, "test-app"),
			KVTag(TagContextPath, "/test-api"),
		))
	}
}

func KVTag(k, v string) string {
	return k + "=" + v
}

type TestServiceRegistration struct {
	tags []string
	meta map[string]interface{}
}

func (r *TestServiceRegistration) ID() string {
	return "dummy-id"
}

func (r *TestServiceRegistration) Name() string {
	return "dummy-name"
}

func (r *TestServiceRegistration) Address() string {
	return "dummy-address"
}

func (r *TestServiceRegistration) Port() int {
	return 9999
}

func (r *TestServiceRegistration) Tags() []string {
	return r.tags
}

func (r *TestServiceRegistration) Meta() map[string]any {
	return r.meta
}

func (r *TestServiceRegistration) SetID(_ string) {
	//noop
}

func (r *TestServiceRegistration) SetName(_ string) {
	//noop
}

func (r *TestServiceRegistration) SetAddress(_ string) {
	//noop
}

func (r *TestServiceRegistration) SetPort(_ int) {
	//noop
}

func (r *TestServiceRegistration) AddTags(tags ...string) {
	r.tags = append(r.tags, tags...)
}

func (r *TestServiceRegistration) RemoveTags(_ ...string) {
	//noop
}

func (r *TestServiceRegistration) SetMeta(key string, value any) {
	if r.meta == nil {
		r.meta = make(map[string]interface{})
	}
	r.meta[key] = value
}
