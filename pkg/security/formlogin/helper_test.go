package formlogin_test

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/access"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/csrf"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/errorhandling"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/request_cache"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	sessioncommon "cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/rest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/sectest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

/*************************
	Mocks
 *************************/

const (
	TestUser1          = `test-user-1`
	TestUser1Password  = `TheCakeIsALie`
	TestUser2          = `test-user-2`
	UrlSecuredEndpoint = `/secured/post`
	UrlGeneralError    = `/error`
	UrlLoginPage       = `/login`
	UrlLoginProcess    = `/login`
	UrlLoginError      = `/login`
	UrlOTPPage         = `/login/mfa`
	UrlOTPProcess      = `/login/otp`
	UrlOTPRefresh      = `/login/otp/refresh`
	UrlOTPError        = `/login/otp`
	ParamUsername      = `username-param`
	ParamPassword      = `password-param`
	ParamCsrf          = `_csrf`
	ParamRememberMe    = `remember-me-param`
	ParamOTP           = `otp`
)

type cfgDI struct {
	fx.In
	SecRegistrar     security.Registrar
	WebRegistrar     *web.Registrar
	AccountStore     *sectest.MockAccountStore
	MFAEventRecorder *MFAEventRecorder `optional:"true"`
}

func FormLoginTestConfigurer(mfa bool) func(di cfgDI) {
	return func(di cfgDI) {
		di.WebRegistrar.MustRegister(TestController{})
		di.SecRegistrar.Register(security.ConfigurerFunc(func(ws security.WebSecurity) {
			ws = ws.Route(matcher.RouteWithPattern("/secured/**")).
				With(access.New().
					Request(matcher.AnyRequest()).Authenticated(),
				).
				With(errorhandling.New().
					AccessDeniedHandler(redirect.NewRedirectWithRelativePath(UrlGeneralError, false)),
				).
				With(session.New()).
				With(csrf.New().
					IgnoreCsrfProtectionMatcher(matcher.RequestWithPattern(UrlSecuredEndpoint)),
				)

			passwdAuth := passwd.Configure(ws).AccountStore(di.AccountStore)
			login := formlogin.Configure(ws).
				LoginUrl(UrlLoginPage).
				LoginProcessUrl(UrlLoginProcess).
				LoginErrorUrl(UrlLoginError).
				UsernameParameter(ParamUsername).
				PasswordParameter(ParamPassword).
				RememberParameter(ParamRememberMe).
				RememberCookieSecured(false).
				RememberCookieDomain("").
				RememberCookieValidity(time.Hour)
			if mfa {
				passwdAuth.MFA(true).
					OtpLength(10).OtpTTL(5 * time.Second).OtpSecretSize(10).
					OtpRefreshLimit(2).OtpVerifyLimit(2)
				if di.MFAEventRecorder != nil {
					passwdAuth.MFAEventListeners(di.MFAEventRecorder.Record)
				}
				login.EnableMFA().
					MfaUrl(UrlOTPPage).
					MfaVerifyUrl(UrlOTPProcess).
					MfaRefreshUrl(UrlOTPRefresh).
					MfaErrorUrl(UrlOTPError).
					OtpParameter(ParamOTP)
			}
		}))
	}
}

type TestController struct{}

func (c TestController) Mappings() []web.Mapping {
	return []web.Mapping{
		rest.Post("/secured/post").EndpointFunc(c.Post).Build(),
	}
}

func (c TestController) Post(_ context.Context, _ *http.Request) (interface{}, error) {
	return map[string]interface{}{
		"success": "yay",
	}, nil
}

type MockedUserAuth struct {
	security.Authentication
	OTP string
}

func (a MockedUserAuth) Username() string {
	return a.Authentication.Principal().(string)
}

func (a MockedUserAuth) IsMFAPending() bool {
	return len(a.OTP) != 0
}

func (a MockedUserAuth) OTPIdentifier() string {
	return a.OTP
}

type MFAEnabledAccount struct {
	security.Account
}

func (MFAEnabledAccount) UseMFA() bool {
	return true
}

type MFAEventRecord struct {
	Event     passwd.MFAEvent
	OTP       passwd.OTP
	Principal interface{}
}

type MFAEventRecorder struct {
	Records []MFAEventRecord
}

func (r *MFAEventRecorder) Record(event passwd.MFAEvent, otp passwd.OTP, principal interface{}) {
	r.Records = append(r.Records, MFAEventRecord{
		Event:     event,
		OTP:       otp,
		Principal: principal,
	})
}

func (r *MFAEventRecorder) Last() MFAEventRecord {
	if idx := len(r.Records) - 1; idx >= 0 {
		return r.Records[idx]
	}
	return MFAEventRecord{}
}

func (r *MFAEventRecorder) Reset() {
	r.Records = make([]MFAEventRecord, 0, 10)
}

func NewMockedAccountStore(accts sectest.MockedPropertiesAccounts) *sectest.MockAccountStore {
	return sectest.NewMockedAccountStore(accts.Values())
}

func NewMockedAccountStoreWithMFA(accts sectest.MockedPropertiesAccounts) *sectest.MockAccountStore {
	return sectest.NewMockedAccountStore(accts.Values(), func(acct security.Account) security.Account {
		return MFAEnabledAccount{Account: acct}
	})
}

func NewMFAEventRecorder() *MFAEventRecorder {
	return &MFAEventRecorder{
		Records: make([]MFAEventRecord, 0, 10),
	}
}

func NewDefaultLoginFormController() *formlogin.DefaultFormLoginController {
	return formlogin.NewDefaultLoginFormController(func(opts *formlogin.DefaultFormLoginPageOptions) {
		opts.LoginTemplate = "test.tmpl"
		opts.LoginProcessUrl = UrlLoginProcess
		opts.UsernameParam = ParamUsername
		opts.PasswordParam = ParamPassword
		opts.MfaTemplate = "test.tmpl"
		opts.MfaVerifyUrl = UrlOTPProcess
		opts.MfaRefreshUrl = UrlOTPRefresh
		opts.OtpParam = ParamOTP
	})
}

// ContextWithLoginPreparation prepare context with CSRF, session and saved request
func ContextWithLoginPreparation(ctx context.Context, store session.Store) (context.Context, *session.Session, string) {
	// mock CSRF token
	csrfToken := utils.RandomString(10)
	ctx, s := ContextWithCsrfToken(ctx, store, csrfToken)
	// mock saved request
	gc := webtest.NewGinContext(ctx, http.MethodPost, UrlSecuredEndpoint, nil)
	request_cache.SaveRequest(gc)
	return ctx, s, csrfToken
}

func ContextWithCsrfToken(ctx context.Context, store session.Store, csrfToken string) (context.Context, *session.Session) {
	return ContextWithSession(ctx, store, csrf.SessionKeyCsrfToken, &csrf.Token{
		Value: csrfToken,
	})
}

func ContextWithSession(ctx context.Context, store session.Store, kvs ...interface{}) (context.Context, *session.Session) {
	const name = sessioncommon.DefaultName
	s, e := store.New(name)
	if e != nil {
		s = session.NewSession(store, name)
	}
	for i := 1; i < len(kvs); i++ {
		s.Set(kvs[i-1], kvs[i])
	}
	ctx = utils.MakeMutableContext(ctx)
	session.MustSet(ctx, s)
	return ctx, s
}

/*************************
	Helper
 *************************/

func WithContextPath(ctx context.Context, path string) string {
	return webtest.CurrentContextPath(ctx) + path
}

func AddSessionCookie(s *session.Session) webtest.RequestOptions {
	return func(req *http.Request) {
		if s == nil {
			return
		}
		cookie := session.NewCookie(s.Name(), s.GetID(), &session.Options{}, req)
		req.AddCookie(cookie)
	}
}

func NewFormLoginRequest(ctx context.Context, s *session.Session, params ...string) *http.Request {
	return NewFormRequestWithSession(ctx, s, http.MethodPost, UrlLoginProcess, params...)
}

func NewOTPVerifyRequest(ctx context.Context, s *session.Session, params ...string) *http.Request {
	return NewFormRequestWithSession(ctx, s, http.MethodPost, UrlOTPProcess, params...)
}

func NewOTPRefreshRequest(ctx context.Context, s *session.Session, params ...string) *http.Request {
	return NewFormRequestWithSession(ctx, s, http.MethodPost, UrlOTPRefresh, params...)
}

func NewFormRequestWithSession(ctx context.Context, s *session.Session, method, path string, params ...string) *http.Request {
	body := url.Values{}
	if len(params) != 0 {
		for i := 1; i < len(params); i += 2 {
			body.Set(params[i-1], params[i])
		}
	}
	headers := []string{"Content-Type", "application/x-www-form-urlencoded"}
	return webtest.NewRequest(ctx, method, path, strings.NewReader(body.Encode()),
		AddSessionCookie(s), webtest.Headers(headers...))
}

func AssertRedirectResponse(g *gomega.WithT, resp *http.Response, expectedLoc string) {
	AssertResponse(g, resp, http.StatusFound)
	g.Expect(resp.Header.Get("Location")).To(Equal(expectedLoc), "response should have correct '%s' header", "Location")
}

func AssertResponse(g *gomega.WithT, resp *http.Response, expectedSC int) []byte {
	g.Expect(resp).ToNot(BeNil(), "response should not be nil")
	g.Expect(resp.StatusCode).To(Equal(expectedSC), "response status code should be correct")
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), "reading response body should not fail")
	return body
}

func AssertCookie(g *gomega.WithT, resp *http.Response, name string) *http.Cookie {
	g.Expect(resp.Cookies()).ToNot(BeEmpty(), "response should have cookies")
	var cookie *http.Cookie
	for _, c := range resp.Cookies() {
		if c.Name == name {
			cookie = c
			break
		}
	}
	if len(name) != 0 {
		g.Expect(cookie).ToNot(BeNil(), "response should have cookie with name '%s'", name)
	} else {
		g.Expect(cookie).To(BeNil(), "response should not have cookie with name '%s'", name)
	}
	return cookie
}

func AssertSession(g *gomega.WithT, resp *http.Response, store session.Store, keys ...interface{}) *session.Session {
	const name = sessioncommon.DefaultName
	cookie := AssertCookie(g, resp, name)
	s, e := store.Get(cookie.Value, name)
	g.Expect(e).To(Succeed(), "get session with ID [%s] and name [%s] should not fail", cookie.Value, name)
	g.Expect(s).ToNot(BeNil(), "session with ID [%s] and name [%s] should not be nil", cookie.Value, name)
	for _, k := range keys {
		g.Expect(s.Get(k)).ToNot(BeZero(), "session should have [%s]", k)
	}
	return s
}

func AssertFlash(g *gomega.WithT, resp *http.Response, store session.Store, keys ...string) *session.Session {
	const name = sessioncommon.DefaultName
	cookie := AssertCookie(g, resp, name)
	s, e := store.Get(cookie.Value, name)
	g.Expect(e).To(Succeed(), "get session with ID [%s] and name [%s] should not fail", cookie.Value, name)
	g.Expect(s).ToNot(BeNil(), "session with ID [%s] and name [%s] should not be nil", cookie.Value, name)
	for _, k := range keys {
		g.Expect(s.Flash(k)).ToNot(BeZero(), "session should have [%s]", k)
	}
	return s
}
