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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cisco-open/go-lanai/pkg/utils/validation"
	"github.com/cisco-open/go-lanai/pkg/web"
	"github.com/cisco-open/go-lanai/pkg/web/matcher"
	"github.com/cisco-open/go-lanai/pkg/web/middleware"
	"github.com/cisco-open/go-lanai/pkg/web/rest"
	"github.com/cisco-open/go-lanai/pkg/web/weberror"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	en_trans "github.com/go-playground/validator/v10/translations/en"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
)

/*************************
	Tests
 *************************/

func TestMvcErrorHandling(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestBindingError(&di), "TestBindingError"),
		test.GomegaSubTest(SubTestHTTPError(&di), "TestHTTPError"),
		test.GomegaSubTest(SubTestGenericError(&di), "TestGenericError"),
		test.GomegaSubTest(SubTestBadRequestError(&di), "TestBadRequestError"),
		test.GomegaSubTest(SubTestCustomError(&di), "TestCustomError"),
		test.GomegaSubTest(SubTestErrorTranslator(&di), "TestErrorTranslator"),
		test.GomegaSubTest(SubTestErrorTranslateMapping(&di), "TestErrorTranslateMapping"),
	)
}

func TestMWErrorHandling(t *testing.T) {
	var di TestDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithUtilities(),
		apptest.WithDI(&di),
		apptest.WithFxOptions(
			fx.Provide(web.NewEngine),
		),
		test.SubTestSetup(ResetEngine(&di)),
		test.GomegaSubTest(SubTestMWGenericError(&di), "TestMWGenericError"),
		test.GomegaSubTest(SubTestMWCustomError(&di), "TestMWCustomError"),
		test.GomegaSubTest(SubTestMWCustomTextError(&di), "TestMWCustomTextError"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestBindingError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// execute test
		WebInit(ctx, t, g, di,
			func(reg *web.Registrar) {
				mappings := []web.Mapping{
					rest.Post("/picky").EndpointFunc(PickyEndpoint).Build(),
				}
				reg.MustRegister(mappings)
				// custom validator and translation
				e := web.Validator().RegisterValidationCtx("custom", validation.Regex("^[z]*$"))
				g.Expect(e).To(Succeed(), "registering custom validation should not fail")
				trans := validation.DefaultTranslator()
				e = web.Validator().SetTranslations(trans, validation.SimpleTranslationRegFunc("custom", `{0} is bad`))
				g.Expect(e).To(Succeed(), "registering translation for custom validation should not fail")
				// default translation
				e = web.Validator().SetTranslations(trans, en_trans.RegisterDefaultTranslations)
				g.Expect(e).To(Succeed(), "setting validator translations should no fail")
				reg.MustRegister(weberror.New("validation").
					ApplyTo(matcher.RouteWithPattern("/picky")).
					Use(testBindingErrorTranslator()).
					Build())
			},
		)
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", HavePrefix(`JsonInt`), func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`{"string":"zzzzz","int":1}`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", `JsonString is bad`, func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`{"string":"not good enough","int":65536}`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", HavePrefix(`invalid character`), func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`bad json`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", HavePrefix(`json:`), func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`"valid but not quite right"`))
		})
	}
}

func SubTestHTTPError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		httpE := web.NewHttpError(http.StatusNotAcceptable, errors.New(DefaultErrorMsg), http.Header{
			ErrorHeaderKey: []string{ErrorHeaderValue},
		})
		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodGet, "/error", httpE),
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error", expectErrorSC(http.StatusNotAcceptable))
	}
}

func SubTestGenericError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := errors.New(DefaultErrorMsg)
		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodDelete, "/error", err),
		)
		testErrorEndpoint(ctx, t, g, http.MethodDelete, "/error", expectErrorHeader(nil))
	}
}

func SubTestBadRequestError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		bre := web.NewBadRequestError(errors.New(DefaultErrorMsg))
		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodGet, "/error", bre),
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error",
			expectErrorSC(http.StatusBadRequest),
			expectErrorHeader(nil),
		)
	}
}

func SubTestCustomError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := customError{
			customErrorBase{
				error: errors.New(DefaultErrorMsg),
				SC:    http.StatusUnauthorized,
			},
		}
		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodGet, "/error", err),
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error", expectErrorSC(err.SC))
	}
}

func SubTestErrorTranslator(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e1 := errors.New("error 1")
		translator1 := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			if err == e1 {
				return fmt.Errorf(`%d`, http.StatusUnauthorized)
			}
			return err
		})
		e2 := errors.New("error 2")
		translator2 := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			if err == e2 {
				return fmt.Errorf(`%d`, http.StatusForbidden)
			}
			return err
		})
		scTranslator := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			sc, _ := strconv.Atoi(err.Error())
			return web.NewHttpError(sc, errors.New(DefaultErrorMsg), http.Header{
				ErrorHeaderKey: []string{ErrorHeaderValue},
			})
		})

		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodGet, "/error/1", e1),
			registerErrorEndpoint(http.MethodGet, "/error/2", e2),
			func(reg *web.Registrar) {
				e := reg.Register(translator1, translator2, scTranslator)
				g.Expect(e).To(Succeed(), "registering error translators should succeed")
			},
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/1", expectErrorSC(http.StatusUnauthorized))
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/2", expectErrorSC(http.StatusForbidden))
	}
}

func SubTestErrorTranslateMapping(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := errors.New("internal error")
		translator1 := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			return fmt.Errorf(`%d`, http.StatusUnauthorized)
		})
		translator1X := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			return fmt.Errorf(`%d`, http.StatusNotAcceptable)
		})
		translator2 := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			return fmt.Errorf(`%d`, http.StatusForbidden)
		})
		scTranslator := web.ErrorTranslateFunc(func(ctx context.Context, err error) error {
			assertContext(ctx, t, g)
			sc, _ := strconv.Atoi(err.Error())
			return web.NewHttpError(sc, errors.New(DefaultErrorMsg), http.Header{
				ErrorHeaderKey: []string{ErrorHeaderValue},
			})
		})

		// execute test
		WebInit(ctx, t, g, di,
			registerErrorEndpoint(http.MethodGet, "/error/1", err),
			registerErrorEndpoint(http.MethodGet, "/error/2", err),
			registerErrorEndpoint(http.MethodPost, "/error/1", err),
			func(reg *web.Registrar) {
				emX := weberror.New("eX").
					ApplyTo(matcher.AnyRoute()).
					Use(scTranslator).Build()
				em1 := weberror.New("e1").Order(-100).
					ApplyTo(matcher.RouteWithPattern("/error/1")).Use(translator1).Build()
				em1X := weberror.New("e1").Order(-10).
					ApplyTo(matcher.RouteWithPattern("/error/1")).
					WithCondition(matcher.RequestWithMethods(http.MethodPost)).
					Use(translator1X).Build()
				em2 := weberror.New("e2").Order(-100).
					ApplyTo(matcher.RouteWithPattern("/error/2")).Use(translator2).Build()
				e := reg.Register(emX, em1, em2, em1X)
				g.Expect(e).To(Succeed(), "registering error translate mappings should succeed")
			},
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/1", expectErrorSC(http.StatusUnauthorized))
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/2", expectErrorSC(http.StatusForbidden))
		testErrorEndpoint(ctx, t, g, http.MethodPost, "/error/1", expectErrorSC(http.StatusNotAcceptable))
	}
}

func SubTestMWGenericError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := errors.New(DefaultErrorMsg)
		// execute test
		WebInit(ctx, t, g, di,
			registerSuccessEndpoint(http.MethodPost, "/mw/error"),
			registerErrorMW(http.MethodPost, "/mw/**", err),
			func(reg *web.Registrar) {
				reg.MustRegister(web.NewRecoveryCustomizer())
				reg.MustRegister(web.NewGinErrorHandlingCustomizer())
			},
		)
		testErrorMW(ctx, t, g, http.MethodPost, "/mw/error", func(expect *errExpectation) {
			expect.bodyDecoder = textBodyDecoder()
			expect.body = map[string]interface{}{
				ErrorBodyKeyMsg: DefaultErrorMsg,
			}
		})
	}
}

func SubTestMWCustomError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := customError{
			customErrorBase{
				error: errors.New(DefaultErrorMsg),
				SC:    http.StatusUnauthorized,
			},
		}
		// execute test
		WebInit(ctx, t, g, di,
			registerSuccessEndpoint(http.MethodPost, "/mw/error"),
			registerErrorMW(http.MethodPost, "/mw/**", err),
			func(reg *web.Registrar) {
				reg.MustRegister(web.NewRecoveryCustomizer())
				reg.MustRegister(web.NewGinErrorHandlingCustomizer())
			},
		)
		testErrorMW(ctx, t, g, http.MethodPost, "/mw/error", expectErrorSC(http.StatusUnauthorized))
	}
}

func SubTestMWCustomTextError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := customTextError{
			customErrorBase{
				error: errors.New(DefaultErrorMsg),
				SC:    http.StatusForbidden,
			},
		}
		// execute test
		WebInit(ctx, t, g, di,
			registerSuccessEndpoint(http.MethodPost, "/mw/error"),
			registerErrorMW(http.MethodPost, "/mw/**", err),
			func(reg *web.Registrar) {
				reg.MustRegister(web.NewRecoveryCustomizer())
				reg.MustRegister(web.NewGinErrorHandlingCustomizer())
			},
		)
		testErrorMW(ctx, t, g, http.MethodPost, "/mw/error",
			expectErrorSC(http.StatusForbidden),
			func(expect *errExpectation) {
				expect.bodyDecoder = urlencodedBodyDecoder()
			},
		)
	}
}

/*************************
	Helper
 *************************/

func testBindingErrorTranslator() web.ErrorTranslateFunc {
	return func(ctx context.Context, err error) error {
		var bindingErr web.BindingError
		var verr validator.ValidationErrors
		var jsonErr *json.SyntaxError
		switch {
		case errors.As(err, &verr):
			// just return first error
			for _, fe := range verr {
				return web.NewBindingError(errors.New(fe.Translate(validation.DefaultTranslator())))
			}
			return err
		case errors.As(err, &jsonErr):
			return  web.NewBindingError(jsonErr)
		case errors.As(err, &bindingErr):
			return bindingErr
		default:
			return err
		}
	}
}

func registerErrorEndpoint(method, path string, err error) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(rest.New(path).
			Method(method).
			Path(path).
			EndpointFunc(errorEndpointFunc(err)).
			Build())
	}
}

func registerErrorMW(method, pattern string, err error) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(middleware.NewBuilder("mw").
			ApplyTo(matcher.RouteWithPattern(pattern, method)).
			Use(errorGinMWFunc(err)).
			Build())
	}
}

func errorEndpointFunc(err error) web.MvcHandlerFunc {
	return func(ctx context.Context, req *http.Request) (interface{}, error) {
		return nil, err
	}
}

func errorGinMWFunc(err error) gin.HandlerFunc {
	return func(gc *gin.Context) {
		_ = gc.Error(err)
		gc.Abort()
	}
}

func expectErrorSC(sc int) func(expect *errExpectation) {
	return func(expect *errExpectation) {
		expect.status = sc
		expect.body[ErrorBodyKeyError] = http.StatusText(sc)
	}
}

func expectErrorHeader(header http.Header) func(expect *errExpectation) {
	return func(expect *errExpectation) {
		expect.headers = map[string]string{}
		for k, v := range header {
			expect.headers[k] = strings.Join(v, " ")
		}
	}
}

func testBindingErrorEndpoint(ctx context.Context, t *testing.T, g *gomega.WithT, method, path string, expectedMsg interface{}, opts ...webtest.RequestOptions) {
	resp := invokeEndpoint(ctx, t, g, method, path, opts...)
	expect := errExpectation{
		status:      http.StatusBadRequest,
		bodyDecoder: jsonBodyDecoder(),
		body: map[string]interface{}{
			"error": "Bad Request",
			"message": expectedMsg,
		},
	}
	assertErrorResponse(t, g, resp, expect)
}

func testErrorMW(ctx context.Context, t *testing.T, g *gomega.WithT, method, path string, expects ...func(expect *errExpectation)) {
	expectOpt := func(expect *errExpectation) {
		expect.headers = nil
	}
	expects = append([]func(expect *errExpectation){expectOpt}, expects...)
	testErrorEndpoint(ctx, t, g, method, path, expects...)
}

func textBodyDecoder() bodyDecoder {
	return func(body io.Reader) (interface{}, error) {
		data, e := io.ReadAll(body)
		if e != nil {
			return nil, e
		}
		return map[string]interface{}{
			ErrorBodyKeyMsg: string(data),
		}, nil
	}
}

/*************************
	Helper Types
 *************************/

type customErrorBase struct {
	error
	SC int
}

func (err customErrorBase) StatusCode() int {
	return err.SC
}

func (err customErrorBase) Headers() http.Header {
	return http.Header{
		ErrorHeaderKey: []string{ErrorHeaderValue},
	}
}

type customError struct {
	customErrorBase
}

func (err customError) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"error":"%s","message":"%s"}`, http.StatusText(err.SC), err.Error())
	return []byte(str), nil
}

type customTextError struct {
	customErrorBase
}

func (err customTextError) MarshalText() ([]byte, error) {
	values := url.Values{}
	values.Set(ErrorBodyKeyError, http.StatusText(err.SC))
	values.Set(ErrorBodyKeyMsg, err.Error())
	return []byte(values.Encode()), nil
}

type PickyRequest struct {
	JsonString string `json:"string" binding:"custom"`
	JsonInt    int    `json:"int" binding:"min=65535"`
}

func PickyEndpoint(_ context.Context, _ *PickyRequest) (interface{}, error) {
	return "not happy", nil
}
