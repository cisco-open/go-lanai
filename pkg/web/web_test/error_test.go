package web_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/validation"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/weberror"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"
)

/*************************
	Tests
 *************************/

func TestErrorResponse(t *testing.T) {
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
		test.GomegaSubTest(SubTestCustomError(&di), "TestCustomError"),
		test.GomegaSubTest(SubTestErrorTranslator(&di), "TestErrorTranslator"),
		test.GomegaSubTest(SubTestErrorTranslateMapping(&di), "TestErrorTranslateMapping"),
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
				_ = web.Validator().RegisterValidationCtx("custom", validation.Regex("^[z]*$"))
			},
		)
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`{"string":"zzzzz","int":1}`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`{"string":"not good enough","int":65536}`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", func(req *http.Request) {
			req.Body = io.NopCloser(strings.NewReader(`bad json`))
		})
		testBindingErrorEndpoint(ctx, t, g, http.MethodPost, "/picky", func(req *http.Request) {
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

func SubTestCustomError(di *TestDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		err := customError{
			error: errors.New(DefaultErrorMsg),
			SC:    http.StatusUnauthorized,
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
			func(reg *web.Registrar) {
				e := reg.Register(scTranslator)
				g.Expect(e).To(Succeed(), "registering error translators should succeed")
				em1 := weberror.New("e1").
					ApplyTo(matcher.RouteWithPattern("/error/1")).Use(translator1).Build()
				em2 := weberror.New("e2").
					ApplyTo(matcher.RouteWithPattern("/error/2")).Use(translator2).Build()
				e = reg.Register(em1, em2)
				g.Expect(e).To(Succeed(), "registering error translate mappings should succeed")
			},
		)
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/1", expectErrorSC(http.StatusUnauthorized))
		testErrorEndpoint(ctx, t, g, http.MethodGet, "/error/2", expectErrorSC(http.StatusForbidden))
	}
}

/*************************
	Helper
 *************************/

func registerErrorEndpoint(method, path string, err error) WebInitFunc {
	return func(reg *web.Registrar) {
		reg.MustRegister(rest.New(path).
			Method(method).
			Path(path).
			EndpointFunc(errorEndpointFunc(err)).
			Build())
	}
}

func errorEndpointFunc(err error) web.MvcHandlerFunc {
	return func(ctx context.Context, req *http.Request) (interface{}, error) {
		return nil, err
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

func testBindingErrorEndpoint(ctx context.Context, t *testing.T, g *gomega.WithT, method, path string, opts ...webtest.RequestOptions) {
	resp := invokeEndpoint(ctx, t, g, method, path, opts...)
	expect := errExpectation{
		status: http.StatusBadRequest,
		bodyDecoder: jsonBodyDecoder(),
	}
	assertErrorResponse(t, g, resp, expect)
}

type customError struct {
	error
	SC int
}

func (err customError) MarshalJSON() ([]byte, error) {
	str := fmt.Sprintf(`{"error":"%s","message":"%s"}`, http.StatusText(err.SC), err.Error())
	return []byte(str), nil
}

func (err customError) StatusCode() int {
	return err.SC
}

func (err customError) Headers() http.Header {
	return http.Header{
		ErrorHeaderKey: []string{ErrorHeaderValue},
	}
}

type PickyRequest struct {
	JsonString string `json:"string" binding:"custom"`
	JsonInt    int    `json:"int" binding:"min=65535"`
}

func PickyEndpoint(_ context.Context, _ *PickyRequest) (interface{}, error) {
	return "not happy", nil
}
