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

package swagger

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/redis"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/config/resserver"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/embedded"
    "github.com/cisco-open/go-lanai/test/sectest"
    "github.com/cisco-open/go-lanai/test/suitetest"
    "github.com/cisco-open/go-lanai/test/webtest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "net/http"
    "testing"
)

func TestMain(m *testing.M) {
	suitetest.RunTests(m,
		embedded.Redis(),
	)
}

func TestSwaggerDocSecurityDisabledWithMockedServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			resserver.Module,
			redis.Module,
			access.Module,
			errorhandling.Module,
		),
		apptest.WithProperties("swagger.security.secure-docs=false", "swagger.spec: testdata/api-docs-v3.yml"),
		apptest.WithFxOptions(
			fx.Provide(
				NewResServerConfigurer,
				bindSwaggerProperties,
			),
			fx.Invoke(
				initialize,
				configureSecurity,
			),
		),
		test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *gomega.WithT) {
			var req *http.Request
			var resp *http.Response
			uri := fmt.Sprintf("http://cisco.com/test/v3/api-docs")
			req = webtest.NewRequest(ctx, http.MethodGet, uri, nil, func(req *http.Request) {
				req.Header.Add("content-type", "application/json")
			})
			resp = webtest.MustExec(ctx, req).Response
			fmt.Printf("%v\n", resp)
			g.Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}, "TestSwaggerDocApiSecurityDisabled"),
	)
}

func TestSwaggerDocSecurityEnabledWithMockedServer(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(
			resserver.Module,
			redis.Module,
			access.Module,
			errorhandling.Module,
		),
		apptest.WithProperties("swagger.security.secure-docs=true", "swagger.spec: testdata/api-docs-v3.yml"),
		apptest.WithFxOptions(
			fx.Provide(
				NewResServerConfigurer,
				bindSwaggerProperties,
			),
			fx.Invoke(
				initialize,
				configureSecurity,
			),
		),
		test.GomegaSubTest(func(ctx context.Context, t *testing.T, g *gomega.WithT) {
			var req *http.Request
			var resp *http.Response
			uri := fmt.Sprintf("http://cisco.com/test/v3/api-docs")
			req = webtest.NewRequest(ctx, http.MethodGet, uri, nil, func(req *http.Request) {
				req.Header.Add("content-type", "application/json")
			})
			resp = webtest.MustExec(ctx, req).Response
			fmt.Printf("%v\n", resp)
			g.Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
		}, "TestSwaggerDocApiSecurityEnabled"),
	)
}

func NewResServerConfigurer() resserver.ResourceServerConfigurer {
	return func(config *resserver.Configuration) {
		//do nothing
	}
}
