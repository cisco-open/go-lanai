package openid

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/auth"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/oauth2/jwt"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

/*************************
	Setup Test
 *************************/

func ProvideOpenIDAuthorizeRequestProcessor(issuer security.Issuer, decoder jwt.JwtDecoder) *OpenIDAuthorizeRequestProcessor {
	return NewOpenIDAuthorizeRequestProcessor(func(opt *ARPOption) {
		opt.Issuer = issuer
		opt.JwtDecoder = decoder
	})
}

/*************************
	Test
 *************************/

type ARProcessorDI struct {
	fx.In
	ARProcessor *OpenIDAuthorizeRequestProcessor
	JwtEncoder  jwt.JwtEncoder
	JwtDecoder  jwt.JwtDecoder
}

func TestOpenIDAuthorizeRequestProcessor(t *testing.T) {
	var di ARProcessorDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithFxOptions(
			fx.Provide(
				BindMockingProperties, NewTestIssuer, NewTestAccountStore,
				NewJwkStore, NewJwtEncoder, NewJwtDecoder,
				ProvideOpenIDAuthorizeRequestProcessor,
			),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestProcessWithMinimumParams(&di), "ProcessWithMinimumParams"),
		test.GomegaSubTest(SubTestProcessWithResponseTypes(&di), "ProcessWithResponseTypes"),
		test.GomegaSubTest(SubTestProcessWithoutNonce(&di), "ProcessWithoutNonce"),
		test.GomegaSubTest(SubTestProcessWithDisplay(&di), "ProcessWithDisplay"),
		test.GomegaSubTest(SubTestProcessWithACR(&di), "ProcessWithClaimsRequest"),
		test.GomegaSubTest(SubTestProcessWithMaxAge(&di), "ProcessWithMaxAge"),
		test.GomegaSubTest(SubTestProcessWithPrompt(&di), "ProcessWithPrompt"),
		test.GomegaSubTest(SubTestProcessWithRequestObject(&di), "ProcessWithRequestObject"),
		test.GomegaSubTest(SubTestProcessWithRequestUri(&di), "ProcessWithRequestUri"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestProcessWithMinimumParams(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
			req.ResponseTypes.Add("code")
		})
		AssertProcessor(ctx, g, di, req, true, "minimum params")

		req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
			req.Scopes.Remove("openid")
			req.ResponseTypes.Add("code")
		})
		AssertProcessor(ctx, g, di, req, true, "no 'openid' scope")
	}
}

func SubTestProcessWithResponseTypes(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		supportedRespTypes := []string{
			"code", "code id_token", "token", "token id_token",
		}
		for _, respType := range supportedRespTypes {
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.ResponseTypes = utils.NewStringSet(strings.Split(respType, " ")...)
				req.Parameters[oauth2.ParameterNonce] = "a nonce"
			})
			AssertProcessor(ctx, g, di, req, true, fmt.Sprintf("response type [%s]", respType))
		}
		unsupportedRespTypes := []string{
			"unknown", "code unknown", "token unknown", "unknown id_token",
		}
		for _, respType := range unsupportedRespTypes {
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.ResponseTypes = utils.NewStringSet(strings.Split(respType, " ")...)
			})
			AssertProcessor(ctx, g, di, req, false, fmt.Sprintf("response type [%s]", respType))
		}
	}
}

func SubTestProcessWithoutNonce(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		type respTypeCond struct {
			success bool
			value   string
		}
		respTypeConds := []respTypeCond{
			{value: "code", success: true}, {value: "code id_token", success: true},
			{value: "token", success: false}, {value: "token id_token", success: false},
		}
		for _, cond := range respTypeConds {
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.ResponseTypes = utils.NewStringSet(strings.Split(cond.value, " ")...)
			})
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("response type [%s]", cond.value))
		}
	}
}

func SubTestProcessWithDisplay(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		type displayCond struct {
			success bool
			value   string
		}
		// Note: we ignore display param if it's not supported
		displayConds := []displayCond{
			{value: "page", success: true}, {value: "touch", success: true},
			{value: "popup", success: true}, {value: "wap", success: true},
			{value: "unknown", success: true},
		}
		for _, cond := range displayConds {
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.Parameters[oauth2.ParameterDisplay] = cond.value
			})
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("display mode [%s]", cond.value))
		}
	}
}

func SubTestProcessWithACR(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		const claimsTmpl = `{"id_token":{"acr":{"essential": %v, "values": ["%s"] }}}`
		type acrCond struct {
			success bool
			claims  string
			acrs    []string
		}
		acrConds := []acrCond{
			{success: true, claims: fmt.Sprintf(claimsTmpl, true, ACRValue(3)), acrs: []string{ACRValue(3), ACRValue(2)}},
			{success: true, claims: fmt.Sprintf(claimsTmpl, true, ACRValue(2)), acrs: []string{ACRValue(1), ACRValue(0)}},
			{success: true, claims: fmt.Sprintf(claimsTmpl, false, ACRValue(3)), acrs: []string{ACRValue(2), ACRValue(1)}},
			{success: true, claims: fmt.Sprintf(claimsTmpl, false, ACRValue(4)), acrs: []string{ACRValue(3)}},
			{success: false, claims: `malformed`, acrs: []string{ACRValue(3), ACRValue(2)}},
			{success: false, claims: fmt.Sprintf(claimsTmpl, true, ACRValue(4)), acrs: []string{}},
		}
		for _, cond := range acrConds {
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.Parameters[oauth2.ParameterClaims] = cond.claims
				req.Parameters[oauth2.ParameterACR] = strings.Join(cond.acrs, " ")
			})
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("ACR [%s] and claims [%s]", cond.acrs, cond.claims))
		}
	}
}

func SubTestProcessWithMaxAge(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		type maxAgeCond struct {
			success  bool
			maxAge   string
			authTime time.Time
		}
		now := time.Now()
		maxAgeConds := []maxAgeCond{
			{success: true, maxAge: "100", authTime: now.Add(-10 * time.Second)},
			{success: true, maxAge: "", authTime: now.Add(-10 * time.Second)},
			{success: true, maxAge: "invalid", authTime: now.Add(-10 * time.Second)},
			{success: false, maxAge: "10", authTime: now.Add(-100 * time.Second)},
			{success: false, maxAge: "0", authTime: now.Add(-10 * time.Second)},
			{success: false, maxAge: "-10", authTime: now.Add(-100 * time.Second)},
			{success: false, maxAge: "0", authTime: time.Time{}},
		}
		for _, cond := range maxAgeConds {
			ctx := ctx
			if !cond.authTime.IsZero() {
				userAuth := sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
					opt.Details = map[string]interface{}{
						security.DetailsKeyAuthTime: cond.authTime,
					}
					opt.State = security.StateAuthenticated
				})
				// note: utils.MakeMutableContext(ctx) or gin.Context is required for clearing security
				ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(userAuth))
			}
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.Parameters[oauth2.ParameterMaxAge] = cond.maxAge
			})
			// max age shouldn't cause error
			AssertProcessor(ctx, g, di, req, true, fmt.Sprintf("max age [%s] and auth time [%v]", cond.maxAge, cond.authTime))
			currentAuth := security.Get(ctx)
			if cond.success {
				g.Expect(security.IsFullyAuthenticated(currentAuth)).
					To(BeTrue(), "current auth should be used with max age [%s] and auth time [%v]", cond.maxAge, cond.authTime)
			} else {
				g.Expect(security.IsFullyAuthenticated(currentAuth)).
					To(BeFalse(), "current auth should be cleared with max age [%s] and auth time [%v]", cond.maxAge, cond.authTime)
			}
		}
	}
}

func SubTestProcessWithPrompt(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		type promptCond struct {
			success       bool
			authCleared   bool
			prompt        string
			authenticated bool
			request       webtest.RequestOptions
		}
		promptConds := []promptCond{
			{success: false, authCleared: true, prompt: "none", authenticated: false},
			{success: true, authCleared: false, prompt: "none", authenticated: true},
			{success: true, authCleared: true, prompt: "login", authenticated: false},
			{success: true, authCleared: true, prompt: "login", authenticated: true},
			{success: true, authCleared: true, prompt: "login", authenticated: false, request: webtest.Headers(keyPromptProcessed, "login")},
			{success: true, authCleared: false, prompt: "login", authenticated: true, request: webtest.Headers(keyPromptProcessed, "login")},
			{success: true, authCleared: true, prompt: "consent", authenticated: false},
			{success: true, authCleared: false, prompt: "consent", authenticated: true},
			{success: true, authCleared: true, prompt: "select_account", authenticated: false},
			{success: true, authCleared: false, prompt: "select_account", authenticated: true},
			{success: true, authCleared: true, prompt: "invalid", authenticated: false},
			{success: true, authCleared: false, prompt: "invalid", authenticated: true},
		}
		for _, cond := range promptConds {
			ctx := ctx
			if cond.authenticated {
				userAuth := sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
					opt.State = security.StateAuthenticated
				})
				ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(userAuth))
			}
			ctx= MockGinContext(ctx, cond.request)
			req = NewOpenIDAuthorizeRequest(ctx, func(req *auth.AuthorizeRequest) {
				req.Parameters[oauth2.ParameterPrompt] = cond.prompt
			})
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("prompt [%s] and authenticated [%v]", cond.prompt, cond.authenticated))
			currentAuth := security.Get(ctx)
			if !cond.authCleared {
				g.Expect(security.IsFullyAuthenticated(currentAuth)).
					To(BeTrue(), "current auth should be used with prompt [%s] and authenticated [%v]", cond.prompt, cond.authenticated)
			} else {
				g.Expect(security.IsFullyAuthenticated(currentAuth)).
					To(BeFalse(), "current auth should be cleared prompt [%s] and authenticated [%v]", cond.prompt, cond.authenticated)
			}
			if strings.Contains(cond.prompt, "login") && cond.authenticated && cond.authCleared {
				g.Expect(ctx.(*gin.Context).Request.Header.Get(keyPromptProcessed)).ToNot(BeZero(), "header [%s] should be set with prompt [%s] and authenticated [%v]", keyPromptProcessed, cond.prompt, cond.authenticated)
			}
		}
	}
}

func SubTestProcessWithRequestObject(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		type reqObjCond struct {
			success bool
			obj     string
			uri     string
		}
		reqObj := []reqObjCond{
			{obj: NewRequestObjectJwt(di.JwtEncoder), success: true},
			{obj: NewRequestObjectJwt(di.JwtEncoder), uri: "http://localhost:0/authorize/request", success: false},
			{obj: NewRequestObjectJwt(di.JwtEncoder, ARResponseTypes("token")), success: false},
			{obj: NewRequestObjectJwt(di.JwtEncoder, ARScopes("read", "write")), success: false},
			{obj: "malformed", success: false},
		}
		for _, cond := range reqObj {
			req = auth.NewAuthorizeRequest(func(req *auth.AuthorizeRequest) {
				req.Scopes.Add(oauth2.ScopeOidc)
				req.ResponseTypes.Add("code")
				if len(cond.obj) != 0 {
					req.Parameters[oauth2.ParameterRequestObj] = cond.obj
				}
				if len(cond.uri) != 0 {
					req.Parameters[oauth2.ParameterRequestUri] = cond.uri
				}
			}).WithContext(ctx)
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("req obj [%s], req uri [%s]", cond.obj, cond.uri))
		}
	}
}

func SubTestProcessWithRequestUri(di *ARProcessorDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *auth.AuthorizeRequest
		// setup an authorize request CDN server
		const path = `/authorize/request`
		handler := &RequestUriHandler{Path: path}
		server := httptest.NewServer(handler)
		defer server.Close()
		// do tests
		type reqObjCond struct {
			success bool
			obj     string
			uri     string
		}
		reqObj := []reqObjCond{
			{uri: server.URL + path, obj: NewRequestObjectJwt(di.JwtEncoder), success: true},
			{uri: server.URL + "/not/exists", obj: NewRequestObjectJwt(di.JwtEncoder), success: false},
			{uri: server.URL + path, obj: NewRequestObjectJwt(di.JwtEncoder, ARResponseTypes("token")), success: false},
			{uri: server.URL + path, obj: NewRequestObjectJwt(di.JwtEncoder, ARScopes("read", "write")), success: false},
			{uri: server.URL + path, obj: "malformed", success: false},
		}
		for _, cond := range reqObj {
			req = auth.NewAuthorizeRequest(func(req *auth.AuthorizeRequest) {
				req.Scopes.Add(oauth2.ScopeOidc)
				req.ResponseTypes.Add("code")
				req.Parameters[oauth2.ParameterRequestUri] = cond.uri
			}).WithContext(ctx)
			handler.JWTValue = cond.obj
			AssertProcessor(ctx, g, di, req, cond.success, fmt.Sprintf("req uri [%s], jwt [%s]", cond.uri, cond.obj))
		}
	}
}

/*************************
	Helpers
 *************************/

func MockGinContext(ctx context.Context, opts ...webtest.RequestOptions) *gin.Context {
	return webtest.NewGinContext(ctx, http.MethodGet, "/test", nil, opts...)
}

func ACRValue(lvl int) string {
	return fmt.Sprintf("http://%s%s/loa-%d", IssuerDomain, IssuerPath, lvl)
}

type MockedARProcessChain struct {
	Continued bool
}

func NewMockedARProcessChain() *MockedARProcessChain {
	return &MockedARProcessChain{}
}

func (c *MockedARProcessChain) Next(_ context.Context, request *auth.AuthorizeRequest) (processed *auth.AuthorizeRequest, err error) {
	c.Continued = true
	return request, nil
}

func NewOpenIDAuthorizeRequest(ctx context.Context, opts ...func(req *auth.AuthorizeRequest)) *auth.AuthorizeRequest {
	defaultOpts := []func(req *auth.AuthorizeRequest){
		ARClientID(ClientIDMinor),
		ARResponseTypes("code"),
		ARScopes("read", "write", oauth2.ScopeOidc),
	}
	defaultOpts = append(defaultOpts, opts...)
	return auth.NewAuthorizeRequest(defaultOpts...).WithContext(ctx)
}

func AssertProcessor(ctx context.Context, g *gomega.WithT, di *ARProcessorDI, ar *auth.AuthorizeRequest, expectPass bool, desc string) {
	chain := NewMockedARProcessChain()
	processed, e := di.ARProcessor.Process(ctx, ar, chain)
	if expectPass {
		g.Expect(e).To(Succeed(), "Process() should not fail with %s", desc)
		g.Expect(chain.Continued).To(BeTrue(), "process chain should be continued with %s", desc)
		g.Expect(processed).ToNot(BeNil(), "processed request should not be nil with %s", desc)
	} else {
		g.Expect(e).To(HaveOccurred(), "Process() should fail with %s", desc)
	}
}

func NewRequestObjectJwt(encoder jwt.JwtEncoder, opts ...func(req *auth.AuthorizeRequest)) string {

	ar := NewOpenIDAuthorizeRequest(context.Background(), opts...)
	claims := oauth2.MapClaims{
		oauth2.ParameterClientId:     ar.ClientId,
		oauth2.ParameterResponseType: strings.Join(ar.ResponseTypes.Values(), " "),
		oauth2.ParameterScope:        strings.Join(ar.Scopes.Values(), " "),
		oauth2.ParameterRedirectUri:  ar.RedirectUri,
		oauth2.ParameterState:        ar.State,
	}
	for k, v := range ar.Parameters {
		claims.Set(k, v)
	}
	for k, v := range ar.Extensions {
		claims.Set(k, v)
	}
	str, _ := encoder.Encode(context.Background(), claims)
	return str
}

type RequestUriHandler struct {
	Path     string
	JWTValue string
}

func (h *RequestUriHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if req.URL.Path != h.Path {
		rw.WriteHeader(http.StatusNotFound)
		return
	}
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write([]byte(h.JWTValue))
}

func ARClientID(value string) func(req *auth.AuthorizeRequest) {
	return func(req *auth.AuthorizeRequest) {
		req.ClientId = value
	}
}

func ARResponseTypes(values ...string) func(req *auth.AuthorizeRequest) {
	return func(req *auth.AuthorizeRequest) {
		req.ResponseTypes = utils.NewStringSet(values...)
	}
}

func ARScopes(values ...string) func(req *auth.AuthorizeRequest) {
	return func(req *auth.AuthorizeRequest) {
		req.Scopes = utils.NewStringSet(values...)
	}
}