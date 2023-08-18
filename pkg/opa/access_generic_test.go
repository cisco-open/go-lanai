package opa_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	opatest "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/test"
	. "cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa/testdata"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Test
 *************************/

func TestGenericAllow(t *testing.T) {
	di := &testDI{}
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		//apptest.WithTimeout(5 * time.Minute),
		opatest.WithBundles(),
		apptest.WithDI(di),
		test.GomegaSubTest(SubTestGenericBaseline(), "TestGenericBaseline"),
		test.GomegaSubTest(SubTestAllowGenericWithAuth(), "TestAllowGenericWithAuth"),
		test.GomegaSubTest(SubTestAllowGenericWithoutAuth(), "TestAllowGenericWithoutAuth"),
		test.GomegaSubTest(SubTestGenericInvalidInputCustomizer(di), "TestGenericInvalidInputCustomizer"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGenericBaseline() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx, func(q *opa.Query) {
			q.Policy = "baseline/allow"
			q.RawInput = map[string]interface{}{
				"just_data": "data",
			}
		})
		g.Expect(e).To(Succeed())
	}
}

func SubTestAllowGenericWithAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx, opa.QueryWithPolicy("baseline/allow_custom"))
		g.Expect(e).To(HaveOccurred())

		ctx = sectest.ContextWithSecurity(ctx, MemberAdminOptions())
		e = opa.Allow(ctx, opa.QueryWithPolicy("baseline/allow_custom"))
		g.Expect(e).To(Succeed())
	}
}

func SubTestAllowGenericWithoutAuth() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		e = opa.Allow(ctx,
			opa.QueryWithPolicy("baseline/allow_custom"),
			opa.QueryWithInputCustomizer(func(ctx context.Context, input *opa.Input) error {
				input.ExtraData["allow_no_auth"] = true
				return nil
			}),
		)
		g.Expect(e).To(Succeed())
	}
}

func SubTestGenericInvalidInputCustomizer(_ *testDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		ctx = sectest.ContextWithSecurity(ctx, MemberOwnerOptions())
		e = opa.Allow(ctx,
			opa.QueryWithPolicy("baseline/allow_custom"),
			opa.QueryWithInputCustomizer(func(ctx context.Context, input *opa.Input) error {
				return errors.New("oops")
			}),
		)
		g.Expect(e).To(HaveOccurred(), "API access should be denied")
		g.Expect(errors.Is(e, opa.ErrInternal)).To(BeTrue(), "error should be ErrInternal")
	}
}
