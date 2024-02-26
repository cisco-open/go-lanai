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

package actuator_tests

import (
	"github.com/cisco-open/go-lanai/pkg/actuator"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
)

/*************************
	Common Test Setup
 *************************/

const (
	SpecialScopeAdmin = "admin"
)

type TestDI struct {
	fx.In
}

/*************************
	Common Helpers
 *************************/

func mockedSecurityAdmin() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Permissions = utils.NewStringSet("IS_API_ADMIN")
	})
}

func mockedSecurityScopedAdmin() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Scopes = utils.NewStringSet(SpecialScopeAdmin)
	})
}

func mockedSecurityNonAdmin() sectest.SecurityContextOptions {
	return sectest.MockedAuthentication(func(d *sectest.SecurityDetailsMock) {
		d.Permissions = utils.NewStringSet("not_worthy")
	})
}

func v3RequestOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Accept", "application/json")
	}
}

func v2RequestOptions() webtest.RequestOptions {
	return func(req *http.Request) {
		req.Header.Set("Accept", actuator.ContentTypeSpringBootV2)
	}
}

func assertResponse(_ *testing.T, g *gomega.WithT, resp *http.Response, expectedStatus int, expectedHeaders ...string) {
	g.Expect(resp).ToNot(BeNil(), "endpoint should have response")
	g.Expect(resp.StatusCode).To(BeEquivalentTo(expectedStatus))
	for i := range expectedHeaders {
		if i%2 == 1 || i+1 >= len(expectedHeaders) {
			continue
		}
		k := expectedHeaders[i]
		v := expectedHeaders[i+1]
		g.Expect(resp.Header.Get(k)).To(BeEquivalentTo(v), "response header should contains [%s]='%s'", k, v)
	}
}

/****************************
	Common Gomega Matchers
 ****************************/
