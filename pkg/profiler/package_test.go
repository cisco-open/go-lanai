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

package profiler

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"net/http"
	"testing"
)

const (
	kContentType      = `Content-Type`
	ContentTypeBinary = `application/octet-stream`
	ContentTypeHTML   = `text/html`
	ContentTypeText   = `text/plain`
)

func TestPProfController(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		apptest.WithModules(Module),
		test.GomegaSubTest(SubTestProfiles(), "TestProfiles"),
		test.GomegaSubTest(SubTestIndex(), "TestIndex"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestProfiles() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		AssertFileDownload(ctx, g, "debug/pprof/goroutine", http.StatusOK)
		AssertFileDownload(ctx, g, "debug/pprof/threadcreate", http.StatusOK)
		AssertFileDownload(ctx, g, "debug/pprof/heap", http.StatusOK)
		AssertFileDownload(ctx, g, "debug/pprof/allocs", http.StatusOK)
		AssertFileDownload(ctx, g, "debug/pprof/block", http.StatusOK)
		AssertFileDownload(ctx, g, "debug/pprof/mutex", http.StatusOK)

        AssertPlainText(ctx, g, "debug/pprof/cmdline", http.StatusOK)
        AssertPlainText(ctx, g, "debug/pprof/symbol", http.StatusOK)

        AssertHTML(ctx, g, "debug/pprof/unknown", http.StatusNotFound)

        // Note: following endpoints are too slow to test
        //AssertPlainText(ctx, g, "debug/pprof/trace", http.StatusOK)
        //AssertPlainText(ctx, g, "debug/pprof/profile", http.StatusOK)
	}
}

func SubTestIndex() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		AssertHTML(ctx, g, "debug/pprof", http.StatusOK)
	}
}

/*************************
	Helpers
 *************************/

func AssertFileDownload(ctx context.Context, g *gomega.WithT, path string, expectedSC int) {
    AssertEndpoint(ctx, g, http.MethodGet, path, expectedSC, ContentTypeBinary)
}

func AssertHTML(ctx context.Context, g *gomega.WithT, path string, expectedSC int) {
	AssertEndpoint(ctx, g, http.MethodGet, path, expectedSC, ContentTypeHTML)
}

func AssertPlainText(ctx context.Context, g *gomega.WithT, path string, expectedSC int) {
    AssertEndpoint(ctx, g, http.MethodGet, path, expectedSC, ContentTypeText)
}

func AssertEndpoint(ctx context.Context, g *gomega.WithT, method, path string, expectedSC int, expectedCT string) {
    req := webtest.NewRequest(ctx, method, path, nil)
    resp := webtest.MustExec(ctx, req).Response
    g.Expect(resp).ToNot(BeNil(), "response should not be nil")
    g.Expect(resp.StatusCode).To(Equal(expectedSC), "[%s %s] should respond correct status code", req.Method, req.URL.Path)
	if expectedSC < 200 || expectedSC > 300 {
		return
	}
    g.Expect(resp.Header.Get(kContentType)).To(HavePrefix(expectedCT), "[%s %s] should respond correct content type", req.Method, req.URL.Path)
}
