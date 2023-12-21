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
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

// AssertEnvResponse fail the test if the response doesn't contain "test" profile.
// This function only support V3 response.
func AssertEnvResponse(t *testing.T, resp *http.Response) {
	g := gomega.NewWithT(t)
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `env response body should be readable`)
	g.Expect(body).To(HaveJsonPathWithValue("$.activeProfiles[0]", "test"), "env response should contains correct active profiles")
	g.Expect(body).To(HaveJsonPath("$.propertySources"), "env response should contains propertySources")
	g.Expect(body).To(HaveJsonPath("$.propertySources[0]"), "env response should contains non-empty propertySources")
}

// AssertAPIListResponse fail the test if the response doesn't contain any "endpoint".
// This function only support V3 response.
func AssertAPIListResponse(t *testing.T, resp *http.Response) {
	g := gomega.NewWithT(t)
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `apilist response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$..endpoint"), "apilist response should contain some endpoint field")
}
