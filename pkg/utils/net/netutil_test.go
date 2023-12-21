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

package netutil

import (
	"github.com/onsi/gomega"
	"net/http"
	"testing"
)

func TestGetForwardedHostName(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://saml.ciscomsx.com/auth/v2/authorize", nil)
	host := GetForwardedHostName(req)

	g := gomega.NewWithT(t)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://saml.ciscomsx.com:443/auth/v2/authorize", nil)
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://192.168.0.1:443/auth/v2/authorize", nil)
	req.Header.Set("X-Forwarded-Host", "saml.ciscomsx.com")
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))

	req, _ = http.NewRequest("GET", "https://192.168.0.1:443/auth/v2/authorize", nil)
	req.Header.Set("X-Forwarded-Host", "saml.ciscomsx.com:443")
	host = GetForwardedHostName(req)
	g.Expect(host).To(gomega.Equal("saml.ciscomsx.com"))
}