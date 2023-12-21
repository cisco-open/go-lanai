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

package actuatortest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/actuator/health"
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

type ExpectedHealthOptions func(h *ExpectedHealth)
type ExpectedHealth struct {
	Status             health.Status
	HasDetails         bool
	HasComponents      bool
	RequiredComponents []string
}

func ExpectHealth(status health.Status) ExpectedHealthOptions {
	return func(h *ExpectedHealth) {
		h.Status = status
	}
}

func ExpectHealthComponents(requiredComps ...string) ExpectedHealthOptions {
	return func(h *ExpectedHealth) {
		h.HasComponents = true
		h.RequiredComponents = requiredComps
	}
}

func ExpectHealthDetails() ExpectedHealthOptions {
	return func(h *ExpectedHealth) {
		h.HasDetails = true
	}
}

// AssertHealthResponse fail the test if given response is not a correct "health" endpoint response.
// By default, this function expect a simple health response with status UP and no details nor components disclosed.
// This function support both V2 and V3 responses, default to V3
func AssertHealthResponse(t *testing.T, resp *http.Response, expectations ...ExpectedHealthOptions) {
	expected := ExpectedHealth{
		Status: health.StatusUp,
	}
	for _, fn := range expectations {
		fn(&expected)
	}

	g := gomega.NewWithT(t)
	// determine response versions
	switch typ := resp.Header.Get("Content-Type"); typ {
	case actuator.ContentTypeSpringBootV2:
		assertHealthResponseV2(t, g, resp, &expected)
	default:
		assertHealthResponseV3(t, g, resp, &expected)
	}
}

func assertHealthResponseV3(_ *testing.T, g *gomega.WithT, resp *http.Response, exp *ExpectedHealth) {
	const jsonPathComponents = "$..components"
	const jsonPathDetails = "$..details"
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `health response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.status", exp.Status.String()), "health response should have status [%v]", exp.Status)

	if exp.HasComponents {
		g.Expect(body).To(HaveJsonPath(jsonPathComponents), "v3 health response should have components")
		for _, comps := range exp.RequiredComponents {
			jsonPath := fmt.Sprintf("$.components.%s", comps)
			g.Expect(body).To(HaveJsonPath(jsonPath), "v3 health response should have '%s' status", comps)
		}
	} else {
		g.Expect(body).NotTo(HaveJsonPath(jsonPathComponents), "v3 health response should not have components")
	}

	if exp.HasDetails {
		g.Expect(body).To(HaveJsonPath(jsonPathDetails), "v3 health response should have details")
	} else {
		g.Expect(body).NotTo(HaveJsonPath(jsonPathDetails), "v3 health response should not have details")
	}
}

func assertHealthResponseV2(_ *testing.T, g *gomega.WithT, resp *http.Response, exp *ExpectedHealth) {
	const jsonPathComponents = "$..details"
	const jsonPathDetails = "$..detailed"
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `health response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.status", ContainElement(exp.Status.String())), "health response should have status [%v]", exp.Status)

	if exp.HasComponents {
		g.Expect(body).To(HaveJsonPath(jsonPathComponents), "v2 health response should have components")
		for _, comps := range exp.RequiredComponents {
			jsonPath := fmt.Sprintf("$.details.%s", comps)
			g.Expect(body).To(HaveJsonPath(jsonPath), "v2 health response should have '%s' status", comps)
		}
	} else {
		g.Expect(body).NotTo(HaveJsonPath(jsonPathComponents), "v2 health response should not have components")
	}

	if exp.HasDetails {
		g.Expect(body).To(HaveJsonPath(jsonPathDetails), "v2 health response should have details")
	} else {
		g.Expect(body).NotTo(HaveJsonPath(jsonPathDetails), "v2 health response should not have details")
	}
}
