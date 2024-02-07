package misc_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth/misc"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tenancy"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/mocks"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http/httptest"
	"testing"
)

/*************************
	Setup Test
 *************************/

const (
	TenantChild1 = "t-1"
	TenantChild2 = "t-2"
	TenantChild11 = "t-1-1"
	TenantChild21 = "t-2-2"
)

func NewMockedTenancyAccessor() tenancy.Accessor {
	return mocks.NewMockTenancyAccessorUsingStrIds([]mocks.TenancyRelationWithStrId{
		{
			ChildId:  TenantChild1,
			ParentId: TestTenantID,
		},
		{
			ChildId:  TenantChild2,
			ParentId: TestTenantID,
		},
		{
			ChildId:  TenantChild11,
			ParentId: TenantChild1,
		},
		{
			ChildId:  TenantChild21,
			ParentId: TenantChild2,
		},
	}, TestTenantID)
}

/*************************
	Test
 *************************/

type THDI struct {
	fx.In
	AuthDI
	Endpoint *misc.TenantHierarchyEndpoint
}

func TestTenantHierarchyEndpoint(t *testing.T) {
	var di THDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(tenancy.Module),
		apptest.WithFxOptions(
			fx.Provide(
				BindMockingProperties, NewTestClientStore, NewMockedTenancyAccessor,
				misc.NewTenantHierarchyEndpoint,
			),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestGetParent(&di), "GetParent"),
		test.GomegaSubTest(SubTestGetChildren(&di), "GetChildren"),
		test.GomegaSubTest(SubTestGetAncestors(&di), "GetAncestors"),
		test.GomegaSubTest(SubTestGetDescendants(&di), "GetDescendants"),
		test.GomegaSubTest(SubTestGetRoot(&di), "GetRoot"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestGetParent(di *THDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var resp string
		var e error
		req := &misc.HierarchyRequest{
			TenantId: TenantChild1,
		}
		// without security
		resp, e = di.Endpoint.GetParent(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// without scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		resp, e = di.Endpoint.GetParent(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// with scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		resp, e = di.Endpoint.GetParent(ctx, req)
		g.Expect(e).To(Succeed(), "endpoint should fail without authentication")
		g.Expect(resp).To(Equal(TestTenantID), "response should be correct")
	}
}

func SubTestGetChildren(di *THDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var resp interface{}
		var e error
		req := &misc.HierarchyRequest{
			TenantId: TenantChild1,
		}
		// without security
		resp, e = di.Endpoint.GetChildren(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// without scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		resp, e = di.Endpoint.GetChildren(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// with scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		resp, e = di.Endpoint.GetChildren(ctx, req)
		g.Expect(e).To(Succeed(), "endpoint should fail without authentication")
		g.Expect(resp).To(ContainElement(TenantChild11), "response should be correct")
	}
}

func SubTestGetAncestors(di *THDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var resp interface{}
		var e error
		req := &misc.HierarchyRequest{
			TenantId: TenantChild11,
		}
		// without security
		resp, e = di.Endpoint.GetAncestors(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// without scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		resp, e = di.Endpoint.GetAncestors(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// with scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		resp, e = di.Endpoint.GetAncestors(ctx, req)
		g.Expect(e).To(Succeed(), "endpoint should fail without authentication")
		g.Expect(resp).To(ContainElements(TenantChild1, TestTenantID), "response should be correct")
	}
}

func SubTestGetDescendants(di *THDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var resp interface{}
		var e error
		req := &misc.HierarchyRequest{
			TenantId: TestTenantID,
		}
		// without security
		resp, e = di.Endpoint.GetDescendants(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// without scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		resp, e = di.Endpoint.GetDescendants(ctx, req)
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// with scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		resp, e = di.Endpoint.GetDescendants(ctx, req)
		g.Expect(e).To(Succeed(), "endpoint should fail without authentication")
		g.Expect(resp).To(ContainElements(TenantChild1, TenantChild2, TenantChild11, TenantChild21), "response should be correct")
	}
}

func SubTestGetRoot(di *THDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var resp string
		var e error
		// without security
		resp, e = di.Endpoint.GetRoot(ctx, &web.EmptyRequest{})
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// without scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDMinor)
		resp, e = di.Endpoint.GetRoot(ctx, &web.EmptyRequest{})
		g.Expect(e).To(HaveOccurred(), "endpoint should fail without authentication")

		// with scope
		ctx = ContextWithClient(ctx, g, &di.AuthDI, ClientIDSuper)
		resp, e = di.Endpoint.GetRoot(ctx, &web.EmptyRequest{})
		g.Expect(e).To(Succeed(), "endpoint should fail without authentication")
		g.Expect(resp).To(Equal(TestTenantID), "response should be correct")

		// verify string encoder
		respWriter := httptest.NewRecorder()
		e = misc.StringResponseEncoder()(ctx, respWriter, resp)
		g.Expect(e).To(Succeed(), "encoding JWT response should not fail")
		g.Expect(respWriter.Header().Get("Content-Type")).To(HavePrefix("application/json"), "JWT response should have correct content-type")
		g.Expect(respWriter.Body.Bytes()).To(Equal([]byte(resp)), "JWT response should have correct body")
	}
}

/*************************
	Helpers
 *************************/

