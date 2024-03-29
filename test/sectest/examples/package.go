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

// Examples. The reason this file exists is to work around an issue existed since go 1.3:
// https://github.com/golang/go/issues/8279
// Note: this issue has been fixed in 1.17

package examples

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/integrate/security/scope"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/access"
    "github.com/cisco-open/go-lanai/pkg/security/basicauth"
    "github.com/cisco-open/go-lanai/pkg/security/errorhandling"
    "github.com/cisco-open/go-lanai/pkg/security/redirect"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "net/http"
)

/*************************
	Examples Setup
 *************************/

type TestTarget struct{}

func (t *TestTarget) DoSomethingWithinSecurityScope(ctx context.Context) error {
	e := scope.Do(ctx, func(scopedCtx context.Context) {
		// scopedCtx contains switched security context
		// do something with scopedCtx...
		_ = t.DoSomethingRequiringSecurity(scopedCtx)
	}, scope.UseSystemAccount())
	return e
}

func (t *TestTarget) DoSomethingRequiringSecurity(ctx context.Context) error {
	auth := security.Get(ctx)
	if !security.IsFullyAuthenticated(auth) {
		return fmt.Errorf("not authenticated")
	}
	return nil
}

const (
	TestSecuredURL    = "/api/v1/secured"
	TestEntryPointURL = "/login"
)

type TestController struct{}

func registerTestController(reg *web.Registrar) {
	reg.MustRegister(&TestController{})
}

func (c *TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.New("secured-get").Get(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
		rest.New("secured-post").Post(TestSecuredURL).
			EndpointFunc(c.Secured).Build(),
	}
}

func (c *TestController) Secured(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"Message": "Yes",
	}, nil
}

type TestSecConfigurer struct{}

func (c *TestSecConfigurer) Configure(ws security.WebSecurity) {
	ws.Route(matcher.RouteWithPattern("/api/**")).
		With(
			basicauth.New().EntryPoint(redirect.NewRedirectWithRelativePath(TestEntryPointURL, false)),
		).
		With(access.New().Request(matcher.AnyRequest()).Authenticated()).
		With(errorhandling.New())
}

func registerTestSecurity(registrar security.Registrar) {
	cfg := TestSecConfigurer{}
	registrar.Register(&cfg)
}