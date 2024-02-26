// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package ittest

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/integrate/httpclient"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
	"github.com/cisco-open/go-lanai/pkg/integrate/security/seclient"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sdtest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"testing"
)

const (
	ServiceNameIDM = `usermanagementgoservice`
)

/*************************
	Tests
 *************************/

//func TestMain(m *testing.M) {
//	suitetest.RunTests(m,
//		PackageHttpRecordingMode(),
//	)
//}

type hcDI struct {
	fx.In
	HttpClient httpclient.Client
}

func TestHttpClientWithoutSecurity(t *testing.T) {
	var di hcDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestHttpClientWithSD(&di), "TestHttpClientWithSD"),
		test.GomegaSubTest(SubTestHttpClientWithoutSD(&di), "TestHttpClientWithoutSD"),
	)
}

type secHcDI struct {
	fx.In
	HttpClient httpclient.Client
	AuthClient seclient.AuthenticationClient
}

func TestHttpClientWithSecurity(t *testing.T) {
	var di secHcDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		WithHttpPlayback(t),
		WithRecordedScopes(),
		sdtest.WithMockedSD(sdtest.DefinitionWithPrefix("mocks.sd")),
		apptest.WithModules(httpclient.Module),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestScopeAndSystemAccount(&di), "TestScopeAndSystemAccount"),
		test.GomegaSubTest(SubTestScopeAndCurrentContext(&di), "TestScopeAndCurrentContext"),
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

func SubTestScopeAndSystemAccount(di *secHcDI) test.GomegaSubTestFunc {
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

func SubTestScopeAndCurrentContext(di *secHcDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// get a real token
		rs, e := di.AuthClient.PasswordLogin(ctx, seclient.WithCredentials("superuser", "superuser"))
		g.Expect(e).To(Succeed(), "initial access token request should be success")

		// Authentication is passed along to the controller in MockedServer mode.
		ctx = sectest.ContextWithSecurity(ctx, sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
			d.AccessToken = rs.Token.Value()
		}))

		var resp *httpclient.Response
		client, e := di.HttpClient.WithService(ServiceNameIDM)
		g.Expect(e).To(Succeed(), "client to svc should be available")

		e = scope.Do(ctx, func(ctx context.Context) {
			resp, e = client.Execute(ctx, httpclient.NewRequest("/api/v8/users/current", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
			assertResponse(t, g, resp, e, http.StatusOK)
		}, scope.WithUsername("admin"))
		g.Expect(e).To(Succeed(), "scope switching should succeed")

		// do it again to trigger cached scopes
		e = scope.Do(ctx, func(ctx context.Context) {
			resp, e = client.Execute(ctx, httpclient.NewRequest("/api/v8/users/current", http.MethodGet), httpclient.CustomResponseDecoder(htmlDecodeFunc()))
			assertResponse(t, g, resp, e, http.StatusOK)
		}, scope.WithUsername("admin"))
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
