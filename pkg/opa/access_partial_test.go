package opa_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/regoexpr"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/open-policy-agent/opa/sdk"
	"testing"
)

/*************************
	Test
 *************************/

func TestFilterResource(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		opatest.WithBundles(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestPartialBaseline(di), "TestPartialBaseline"),
		test.GomegaSubTest(SubTestSuccessfulPartial(di), "TestSuccessfulPartial"),
		test.GomegaSubTest(SubTestDeniedPartial(di), "TestDeniedPartial"),
		test.GomegaSubTest(SubTestPartialWithoutPolicy(di), "TestPartialWithoutPolicy"),
		test.GomegaSubTest(SubTestResourceInvalidInputCustomizer(di), "TestResourceInvalidInputCustomizer"),
		test.GomegaSubTest(SubTestPartialInvalidInputCustomizer(di), "TestPartialInvalidInputCustomizer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestPartialBaseline(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var pq *sdk.PartialResult
		var e error
		pq, e = opa.FilterResource(ctx, "doesn't matter", "whatever", func(res *opa.ResourceFilter) {
			res.Policy = "data.baseline.filter"
			res.Unknowns = []string{"input.fail"}
			res.RawInput = map[string]interface{}{
				"just_data": "data",
			}
			res.QueryMapper = regoexpr.NoopPartialQueryMapper{}
		})
		g.Expect(e).To(Succeed())
		g.Expect(pq).ToNot(BeNil())
	}
}

func SubTestSuccessfulPartial(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var pq *sdk.PartialResult
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		// member admin - can read
		pq, e = opa.FilterResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceFilter) {
			res.Unknowns = []string{
				"input.resource.tenant_id",
				"input.resource.tenant_path",
				"input.resource.owner_id",
			}
			res.ExtraData["debug"] = "test"
			res.QueryMapper = regoexpr.NoopPartialQueryMapper{}
		})
		g.Expect(e).To(Succeed())
		g.Expect(pq).ToNot(BeNil())
	}
}

func SubTestDeniedPartial(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var pq *sdk.PartialResult
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		// member admin - can read
		pq, e = opa.FilterResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceFilter) {
			res.Unknowns = []string{}
			res.ExtraData["debug"] = "test"
			res.QueryMapper = regoexpr.NoopPartialQueryMapper{}
		})
		g.Expect(e).To(HaveOccurred())
		g.Expect(pq).To(BeNil())
	}
}

func SubTestPartialWithoutPolicy(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var pq *sdk.PartialResult
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		// member admin - can read
		pq, e = opa.FilterResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceFilter) {
			res.Unknowns = []string{
				"input.resource.tenant_id",
				"input.resource.tenant_path",
				"input.resource.owner_id",
			}
			res.ExtraData["debug"] = "test"
			res.Policy = "data.poc.unknown_policy"
			res.QueryMapper = regoexpr.NoopPartialQueryMapper{}
		})
		g.Expect(e).To(HaveOccurred())
		g.Expect(errors.Is(e, opa.ErrAccessDenied)).To(BeTrue(), "error should be ErrAccessDenied")
		g.Expect(pq).To(BeNil())
	}
}

func SubTestPartialInvalidInputCustomizer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var pq *sdk.PartialResult
		var e error
		// member admin
		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		// member admin - can read
		pq, e = opa.FilterResource(ctx, "poc", opa.OpWrite, func(res *opa.ResourceFilter) {
			res.Unknowns = []string{
				"input.resource.tenant_id",
				"input.resource.tenant_path",
				"input.resource.owner_id",
			}
			res.ExtraData["debug"] = "test"
			res.QueryMapper = regoexpr.NoopPartialQueryMapper{}
			res.InputCustomizers = append(res.InputCustomizers, opa.InputCustomizerFunc(func(ctx context.Context, input *opa.Input) error {
				return errors.New("oops")
			}))
		})
		g.Expect(e).To(HaveOccurred())
		g.Expect(errors.Is(e, opa.ErrInternal)).To(BeTrue(), "error should be ErrInternal")
		g.Expect(pq).To(BeNil())
	}
}