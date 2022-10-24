package ittest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/discovery"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/httpclient"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/integrate/security/scope"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sdtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

const (
	ServiceNameIDM  = `usermanagementgoservice`
	ServiceNameAuth = `authservice`
)

/*************************
	Tests
 *************************/

type hcDI struct {
	fx.In
	HttpClient httpclient.Client
}

func TestHttpClientWithoutSecurity(t *testing.T) {
	var di hcDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t),
		//WithHttpPlayback(t, EnableHttpRecordMode()),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpClientWithSD(&di), "TestHttpClientWithSD"),
		test.GomegaSubTest(SubTestHttpClientWithoutSD(&di), "TestHttpClientWithoutSD"),
	)
}

func TestHttpClientWithSecurity(t *testing.T) {
	var di hcDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t),
		//WithHttpPlayback(t, EnableHttpRecordMode()),
		WithRecordedScopes(),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(
				discovery.NewCustomizers,
			),
		),
		test.GomegaSubTest(SubTestHttpClientWithScope(&di), "TestHttpClientWithScope"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestHttpClientWithSD(di *hcDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var resp *httpclient.Response
		client, e := di.HttpClient.WithService(ServiceNameIDM)
		g.Expect(e).To(Succeed(), "client to svc should be available")

		resp, e = client.Execute(ctx, httpclient.NewRequest("/swagger", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
		assertResponse(t, g, resp, e, http.StatusOK)
	}
}

func SubTestHttpClientWithoutSD(di *hcDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var resp *httpclient.Response
		client, e := di.HttpClient.WithBaseUrl("http://localhost:9203")
		g.Expect(e).To(Succeed(), "client to svc should be available")

		resp, e = client.Execute(ctx, httpclient.NewRequest("/idm/swagger", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
		assertResponse(t, g, resp, e, http.StatusOK)
	}
}

func SubTestHttpClientWithScope(di *hcDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var e error
		var resp *httpclient.Response
		client, e := di.HttpClient.WithService(ServiceNameIDM)
		g.Expect(e).To(Succeed(), "client to svc should be available")

		e = scope.Do(ctx, func(ctx context.Context) {
			resp, e = client.Execute(ctx, httpclient.NewRequest("/api/v8/users/current", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
			assertResponse(t, g, resp, e, http.StatusOK)
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope switching should succeed")

		// do it again to trigger cached scopes
		e = scope.Do(ctx, func(ctx context.Context) {
			resp, e = client.Execute(ctx, httpclient.NewRequest("/api/v8/users/current", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
			assertResponse(t, g, resp, e, http.StatusOK)
		}, scope.UseSystemAccount())
		g.Expect(e).To(Succeed(), "scope switching should succeed")
	}
}

/*************************
	internal
 *************************/

func htmlDecodeFunc() func(context.Context, *http.Response) (response interface{}, err error) {
	return func(ctx context.Context, resp *http.Response) (body interface{}, err error) {
		raw, e := io.ReadAll(resp.Body)
		if e != nil {
			return nil, e
		}
		return &httpclient.Response{
			StatusCode: resp.StatusCode,
			Headers:    resp.Header,
			Body:       raw,
			RawBody:    raw,
		}, nil
	}
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *httpclient.Response, e error, expectedSC int) {
	g.Expect(e).To(Succeed(), "request to svc should succeed")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response code should be correct")
	g.Expect(resp.Body).To(BeAssignableToTypeOf([]byte{}), "response body should be bytes")
	g.Expect(len(resp.Body.([]byte))).To(BeNumerically(">", 0), "response should not be empty")
}
