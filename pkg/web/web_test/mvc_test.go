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

package web_test

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/rest"
    "github.com/cisco-open/go-lanai/pkg/web/web_test/testdata"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/webtest"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "go.uber.org/fx"
    "net/http"
    "strings"
    "testing"
)

/*************************
	Tests
 *************************/

func TestRestMvcRegistration(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestWithController(&di), "TestWithController"),
		test.GomegaSubTest(SubTestWithControllerStruct(&di), "TestWithControllerStruct"),
		test.GomegaSubTest(SubTestWithMvcMapping(&di), "TestWithMvcMapping"),
		test.GomegaSubTest(SubTestWithMvcVariations(&di), "TestWithMvcVariations"),
		test.GomegaSubTest(SubTestWithInvalidMvcHandler(&di), "TestWithInvalidMvcHandler"),
		test.GomegaSubTest(SubTestWithTextResponse(&di), "TestWithTextResponse"),
		test.GomegaSubTest(SubTestWithBytesResponse(&di), "TestWithBytesResponse"),
		test.GomegaSubTest(SubTestWithCustomResponseEncoder(&di), "TestWithCustomResponseEncoder"),
	)
}

func TestRealServer(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithRealServer(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			web.FxControllerProviders(testdata.NewController),
		),
		test.GomegaSubTest(SubTestRealServerSmokeTest(&di), "RealServerSmokeTest"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestWithController(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			e := reg.Register(testdata.Controller{})
			g.Expect(e).To(Succeed(), "register controller should success")
		})
		testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

type controllers struct {
	fx.In
	TestController testdata.Controller
	unexported     string
}

func SubTestWithControllerStruct(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			e := reg.Register(controllers{
				TestController: testdata.Controller{},
			})
			g.Expect(e).To(Succeed(), "register controller should success")
		})
		testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

func SubTestWithMvcMapping(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			e := reg.Register(rest.Post("/basic/:var").EndpointFunc(testdata.StructPtr200).Build())
			g.Expect(e).To(Succeed(), "register MVC mapping should success")
		})
		testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

func SubTestWithMvcVariations(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const uriVar = `:var`
		variations := map[web.MvcMapping]func(*mvcExpectation){
			rest.Post("/struct/ptr/" + uriVar).EndpointFunc(testdata.StructPtr200).Build(): nil,
			rest.Post("/struct/" + uriVar).EndpointFunc(testdata.Struct200).Build():        nil,
			rest.Post("/struct/ptr/201/" + uriVar).EndpointFunc(testdata.StructPtr201).Build(): func(expect *mvcExpectation) {
				expect.status = http.StatusCreated
			},
			rest.Post("/struct/201/" + uriVar).EndpointFunc(testdata.Struct201).Build(): func(expect *mvcExpectation) {
				expect.status = http.StatusCreated
			},
			rest.Post("/struct/ptr/201/header" + uriVar).EndpointFunc(testdata.StructPtr201WithHeader).Build(): func(expect *mvcExpectation) {
				expect.status = http.StatusCreated
				expect.headers = map[string]string{BasicHeaderKey: BasicHeaderValue}
			},
			rest.Post("/struct/201/header" + uriVar).EndpointFunc(testdata.Struct201WithHeader).Build(): func(expect *mvcExpectation) {
				expect.status = http.StatusCreated
				expect.headers = map[string]string{BasicHeaderKey: BasicHeaderValue}
			},
			rest.Post("/raw/" + uriVar).EndpointFunc(testdata.Raw).Build(): nil,
			rest.Post("/no/request/" + uriVar).EndpointFunc(testdata.NoRequest).Build(): func(expect *mvcExpectation) {
				expect.body = map[string]interface{}{"uri": "", "q": "", "header": "", "string": "", "int": float64(0)}
			},
		}

		// test registration
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			mappings := make([]web.MvcMapping, 0, len(variations))
			for k := range variations {
				mappings = append(mappings, k)
			}
			e := reg.Register(mappings)
			g.Expect(e).To(Succeed(), "register MVC mappings should success")
		})

		// test endpoints
		for k, v := range variations {
			method := k.Method()
			path := strings.ReplaceAll(k.Path(), uriVar, "var-value")
			testEndpoint(ctx, t, g, method, path, v)
		}
	}
}

func SubTestWithInvalidMvcHandler(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {

		variations := []rest.EndpointFunc{
			testdata.MissingResponse,
			testdata.MissingError,
			testdata.MissingContext,
			testdata.WrongErrorPosition,
			testdata.WrongContextPosition,
			testdata.ExtraInput,
			"Not a Func",
		}

		// test registration
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			for _, v := range variations {
				e := withRecover(func() error {
					return reg.Register(rest.Post("/invalid").EndpointFunc(v).Build())
				})
				g.Expect(e).To(HaveOccurred(), "register MVC mapping [%T] should fail", v)
			}
		})
	}
}

func SubTestWithTextResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		mappings := []web.MvcMapping{
			rest.Post("/text/:var").EndpointFunc(testdata.Text).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
			rest.Post("/text/string/:var").EndpointFunc(testdata.TextString).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
			rest.Post("/text/bytes/:var").EndpointFunc(testdata.TextBytes).EncodeResponseFunc(web.TextResponseEncoder()).Build(),
		}

		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			e := reg.Register(mappings)
			g.Expect(e).To(Succeed(), "register MVC mapping should success")
		})

		for _, m := range mappings {
			path := strings.ReplaceAll(m.Path(), ":var", "var-value")
			testEndpoint(ctx, t, g, m.Method(), path, func(expect *mvcExpectation) {
				expect.headers = map[string]string{
					"Content-Type": "text/plain; charset=utf-8",
				}
				expect.bodyDecoder = urlencodedBodyDecoder()
				expect.body["int"] = "20"
			})
		}
	}
}

func SubTestWithBytesResponse(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		mappings := []web.MvcMapping{
			rest.Post("/bytes/:var").EndpointFunc(testdata.Bytes).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
			rest.Post("/bytes/string/:var").EndpointFunc(testdata.BytesString).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
			rest.Post("/bytes/struct/:var").EndpointFunc(testdata.BytesStruct).EncodeResponseFunc(web.BytesResponseEncoder()).Build(),
		}

		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			e := reg.Register(mappings)
			g.Expect(e).To(Succeed(), "register MVC mapping should success")
		})

		for _, m := range mappings {
			path := strings.ReplaceAll(m.Path(), ":var", "var-value")
			testEndpoint(ctx, t, g, m.Method(), path, func(expect *mvcExpectation) {
				expect.headers = map[string]string{
					"Content-Type": "application/octet-stream",
				}
				expect.bodyDecoder = urlencodedBodyDecoder()
				expect.body["int"] = "20"
			})
		}
	}
}

func SubTestWithCustomResponseEncoder(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const customContentType = `application/also-json`
		WebInit(ctx, t, g, di, func(reg *web.Registrar) {
			encoder := web.CustomResponseEncoder(func(opt *web.EncodeOption) {
				opt.ContentType = customContentType
				opt.WriteFunc = web.JsonWriteFunc
			})
			e := reg.Register(rest.Post("/basic/:var").EndpointFunc(testdata.StructPtr200).EncodeResponseFunc(encoder).Build())
			g.Expect(e).To(Succeed(), "register MVC mapping should success")
		})

		testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value", func(expect *mvcExpectation) {
			expect.headers = map[string]string{
				"Content-Type": customContentType,
			}
		})
	}
}

func SubTestRealServerSmokeTest(_ *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		testEndpoint(ctx, t, g, http.MethodPost, "/basic/var-value")
	}
}

/*************************
	Helpers
 *************************/

func withRecover(fn func() error) (err error) {
	defer func() {
		if v := recover(); v != nil {
			err = fmt.Errorf("%v", v)
		}
	}()
	return fn()
}
