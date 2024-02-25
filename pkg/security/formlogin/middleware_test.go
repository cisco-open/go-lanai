package formlogin_test

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/security"
	"github.com/cisco-open/go-lanai/pkg/security/access"
	"github.com/cisco-open/go-lanai/pkg/security/csrf"
	"github.com/cisco-open/go-lanai/pkg/security/errorhandling"
	"github.com/cisco-open/go-lanai/pkg/security/formlogin"
	"github.com/cisco-open/go-lanai/pkg/security/passwd"
	"github.com/cisco-open/go-lanai/pkg/security/redirect"
	"github.com/cisco-open/go-lanai/pkg/security/request_cache"
	"github.com/cisco-open/go-lanai/pkg/security/session"
	"github.com/cisco-open/go-lanai/test"
	"github.com/cisco-open/go-lanai/test/apptest"
	"github.com/cisco-open/go-lanai/test/sectest"
	"github.com/cisco-open/go-lanai/test/webtest"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"go.uber.org/fx"
	"net/http"
	"testing"
	"time"
)

/*************************
	Test Setup
 *************************/

/*************************
	Tests
 *************************/

type LoginDI struct {
	fx.In
	SessionStore     session.Store
	AccountStore     *sectest.MockAccountStore
	MFAEventRecorder *MFAEventRecorder `optional:"true"`
}

func TestMiddlewareWithoutMFA(t *testing.T) {
	var di LoginDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module,
			passwd.Module, formlogin.Module, csrf.Module),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesTenants]("tenants"),
				NewMockedAccountStore,
			),
			fx.Invoke(FormLoginTestConfigurer(false)),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestLoginEntrypoint(&di), "LoginEntrypoint"),
		test.GomegaSubTest(SubTestLoginSuccess(&di), "LoginSuccess"),
		test.GomegaSubTest(SubTestLoginSuccessWithRememberMe(&di), "LoginSuccessWithRememberMe"),
		test.GomegaSubTest(SubTestLoginWrongPassword(&di), "LoginWrongPassword"),
		test.GomegaSubTest(SubTestLoginMissingParams(&di), "LoginMissingParams"),
		test.GomegaSubTest(SubTestLoginSameUser(&di), "LoginSameUser"),
	)
}

func TestMiddlewareWithMFA(t *testing.T) {
	var di LoginDI
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		webtest.WithMockedServer(),
		sectest.WithMockedMiddleware(sectest.MWEnableSession()),
		apptest.WithModules(security.Module, access.Module, errorhandling.Module,
			passwd.Module, formlogin.Module, csrf.Module),
		apptest.WithFxOptions(
			fx.Provide(
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesAccounts]("accounts"),
				sectest.MockedPropertiesBinder[sectest.MockedPropertiesTenants]("tenants"),
				NewMockedAccountStoreWithMFA,
				NewMFAEventRecorder,
			),
			fx.Invoke(FormLoginTestConfigurer(true)),
		),
		apptest.WithDI(&di),
		test.GomegaSubTest(SubTestLoginEntrypoint(&di), "LoginEntrypoint"),
		test.GomegaSubTest(SubTestMFAEntrypoint(&di), "MFAEntrypoint"),
		test.GomegaSubTest(SubTestOTPVerifySuccess(&di), "OTPVerifySuccess"),
		test.GomegaSubTest(SubTestOTPVerifyWrongPasscode(&di), "OTPVerifyWrongPasscode"),
		test.GomegaSubTest(SubTestOTPVerifyMissingParams(&di), "OTPVerifyMissingParams"),
		test.GomegaSubTest(SubTestOTPVerifyReachLimit(&di), "OTPVerifyReachLimit"),
		test.GomegaSubTest(SubTestOTPVerifyWithoutAuth(&di), "OTPVerifyWithoutAuth"),
		test.GomegaSubTest(SubTestOTPRefreshSuccess(&di), "OTPRefreshSuccess"),
		test.GomegaSubTest(SubTestOTPRefreshReachLimit(&di), "OTPRefreshReachLimit"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestLoginEntrypoint(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		req = webtest.NewRequest(ctx, http.MethodPost, UrlSecuredEndpoint, nil)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginPage))
		AssertSession(g, resp, di.SessionStore, csrf.SessionKeyCsrfToken, request_cache.SessionKeyCachedRequest)
	}
}

func SubTestLoginSuccess(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		AssertFormLoginSuccess(ctx, g, di, s, csrfToken)
	}
}

func SubTestLoginSuccessWithRememberMe(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		req := NewFormLoginRequest(ctx, s, ParamUsername,
			TestUser1, ParamPassword, TestUser1Password, ParamCsrf, csrfToken, ParamRememberMe, "true")
		resp := webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, UrlSecuredEndpoint)
		AssertSession(g, resp, di.SessionStore, request_cache.SessionKeyCachedRequest)
		cookie := AssertCookie(g, resp, formlogin.CookieKeyRememberedUsername)
		g.Expect(cookie.Value).To(Equal(TestUser1), "remembed username should be correct")
	}
}

func SubTestLoginWrongPassword(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		req = NewFormLoginRequest(ctx, s, ParamUsername, TestUser1, ParamPassword, "oops", ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginError))
		AssertSession(g, resp, di.SessionStore, "error")
		AssertFlash(g, resp, di.SessionStore, ParamUsername, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)

		// try again with correct password
		AssertFormLoginSuccess(ctx, g, di, s, csrfToken)
	}
}

func SubTestLoginMissingParams(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		req = NewFormLoginRequest(ctx, s, ParamUsername, ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginError))
		AssertSession(g, resp, di.SessionStore, "error")
		AssertFlash(g, resp, di.SessionStore, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)

		// try again with correct password
		AssertFormLoginSuccess(ctx, g, di, s, csrfToken)
	}
}

func SubTestLoginSameUser(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		// mock existing auth
		auth := MockedUserAuth{
			Authentication: sectest.NewMockedUserAuthentication(func(opt *sectest.MockUserAuthOption) {
				opt.Principal = TestUser1
				opt.State = security.StateAuthenticated
			}),
		}
		ctx = sectest.ContextWithSecurity(ctx, sectest.Authentication(auth))

		req = NewFormLoginRequest(ctx, s, ParamUsername, TestUser1, ParamPassword, "oops", ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginError))
		AssertSession(g, resp, di.SessionStore, "error")
		AssertFlash(g, resp, di.SessionStore, ParamUsername, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)

		// try again with correct password
		AssertFormLoginSuccess(ctx, g, di, s, csrfToken)
	}
}

func SubTestMFAEntrypoint(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		req = NewFormLoginRequest(ctx, s, ParamUsername, TestUser1, ParamPassword, TestUser1Password, ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPPage))
		AssertSession(g, resp, di.SessionStore, request_cache.SessionKeyCachedRequest)
		otp := AssertLastOTPEvent(ctx, g, di, passwd.MFAEventOtpCreate, TestUser1)
		g.Expect(otp).ToNot(BeNil(), "created OTP should not be nil")
	}
}

func SubTestOTPVerifySuccess(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)

		AssertOTPVerifySuccess(ctx, g, di, otp.Passcode(), s, csrfToken)
	}
}

func SubTestOTPVerifyWrongPasscode(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)

		req = NewOTPVerifyRequest(ctx, s, ParamOTP, "wrong passcode", ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPError))
		AssertSession(g, resp, di.SessionStore, "error")
		AssertFlash(g, resp, di.SessionStore, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)
		otp = AssertLastOTPEvent(ctx, g, di, passwd.MFAEventVerificationFailure, TestUser1)
		g.Expect(otp.Attempts()).To(BeEquivalentTo(1), "OTP verification failure should be recorded")

		// try again
		AssertOTPVerifySuccess(ctx, g, di, otp.Passcode(), s, csrfToken)
	}
}

func SubTestOTPVerifyMissingParams(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)

		req = NewOTPVerifyRequest(ctx, s, ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPError))
		AssertSession(g, resp, di.SessionStore, "error")
		AssertFlash(g, resp, di.SessionStore, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)
		otp = AssertLastOTPEvent(ctx, g, di, passwd.MFAEventVerificationFailure, TestUser1)
		g.Expect(otp.Attempts()).To(BeEquivalentTo(1), "OTP verification failure should be recorded")

		// try again
		AssertOTPVerifySuccess(ctx, g, di, otp.Passcode(), s, csrfToken)
	}
}

func SubTestOTPVerifyReachLimit(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const limit = 2
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)
		for i := 1; i <= limit; i++ {
			req = NewOTPVerifyRequest(ctx, s, ParamOTP, "wrong", ParamCsrf, csrfToken)
			resp = webtest.MustExec(ctx, req).Response
			AssertSession(g, resp, di.SessionStore, "error")
			AssertFlash(g, resp, di.SessionStore, redirect.FlashKeyPreviousError, redirect.FlashKeyPreviousStatusCode)
			otp = AssertLastOTPEvent(ctx, g, di, passwd.MFAEventVerificationFailure, TestUser1)
			g.Expect(otp.Attempts()).To(BeEquivalentTo(i), "OTP verification failure should be recorded")
			if i != limit {
				AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPError))
			} else {
				// last time should redirect to login page and current auth should be revoked
				AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginError))
				s = AssertSession(g, resp, di.SessionStore, "Security")
				g.Expect(s.Get("Security")).To(BeAssignableToTypeOf(security.EmptyAuthentication("")),
					"security in session should be reset to empty")
			}
		}

		// try again
		req = NewOTPVerifyRequest(ctx, s, ParamOTP, otp.Passcode(), ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlGeneralError))
	}
}

func SubTestOTPVerifyWithoutAuth(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		req = NewOTPVerifyRequest(ctx, s, ParamOTP, "anything", ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlGeneralError))
	}
}

func SubTestOTPRefreshSuccess(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)
		id := otp.ID()
		oldPasscode := otp.Passcode()

		// wait at least one second for refreshed OTP to change
		time.Sleep(time.Second)
		req := NewOTPRefreshRequest(ctx, s, ParamCsrf, csrfToken)
		resp := webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPPage))
		refreshed := AssertLastOTPEvent(ctx, g, di, passwd.MFAEventOtpRefresh, TestUser1)
		g.Expect(refreshed.ID()).To(Equal(id), "refreshed OTP ID shouldn't change")
		g.Expect(refreshed.Passcode()).ToNot(Equal(oldPasscode), "refreshed OTP passcode should change")
		g.Expect(refreshed.Refreshes()).To(BeEquivalentTo(1), "refreshed OTP refresh count should be correct")
	}
}

func SubTestOTPRefreshReachLimit(di *LoginDI) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const limit = 2
		var req *http.Request
		var resp *http.Response
		di.MFAEventRecorder.Reset()
		ctx, s, csrfToken := ContextWithLoginPreparation(ctx, di.SessionStore)
		otp := AssertMFAPasswordSuccess(ctx, g, di, s, csrfToken)
		for i := 1; i <= limit; i++ {
			req = NewOTPRefreshRequest(ctx, s, ParamCsrf, csrfToken)
			resp = webtest.MustExec(ctx, req).Response
			AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPPage))
			g.Expect(otp.Refreshes()).To(BeEquivalentTo(i), "OTP refresh should be recorded")
		}

		// try again
		req = NewOTPRefreshRequest(ctx, s, ParamOTP, otp.Passcode(), ParamCsrf, csrfToken)
		resp = webtest.MustExec(ctx, req).Response
		AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlLoginError))
		s = AssertSession(g, resp, di.SessionStore, "Security")
		g.Expect(s.Get("Security")).To(BeAssignableToTypeOf(security.EmptyAuthentication("")),
			"security in session should be reset to empty")
	}
}

/*************************
	Helpers
 *************************/

func AssertFormLoginSuccess(ctx context.Context, g *gomega.WithT, di *LoginDI, s *session.Session, csrfToken string) {
	req := NewFormLoginRequest(ctx, s, ParamUsername, TestUser1, ParamPassword, TestUser1Password, ParamCsrf, csrfToken)
	resp := webtest.MustExec(ctx, req).Response
	AssertRedirectResponse(g, resp, UrlSecuredEndpoint)
	AssertSession(g, resp, di.SessionStore, request_cache.SessionKeyCachedRequest)
}

func AssertMFAPasswordSuccess(ctx context.Context, g *gomega.WithT, di *LoginDI, s *session.Session, csrfToken string) passwd.OTP {
	req := NewFormLoginRequest(ctx, s, ParamUsername, TestUser1, ParamPassword, TestUser1Password, ParamCsrf, csrfToken)
	resp := webtest.MustExec(ctx, req).Response
	AssertRedirectResponse(g, resp, WithContextPath(ctx, UrlOTPPage))
	AssertSession(g, resp, di.SessionStore, request_cache.SessionKeyCachedRequest)
	return AssertLastOTPEvent(ctx, g, di, passwd.MFAEventOtpCreate, TestUser1)
}

func AssertOTPVerifySuccess(ctx context.Context, g *gomega.WithT, di *LoginDI, passcode string, s *session.Session, csrfToken string) {
	req := NewOTPVerifyRequest(ctx, s, ParamOTP, passcode, ParamCsrf, csrfToken)
	resp := webtest.MustExec(ctx, req).Response
	AssertRedirectResponse(g, resp, UrlSecuredEndpoint)
	AssertSession(g, resp, di.SessionStore, request_cache.SessionKeyCachedRequest)
	AssertLastOTPEvent(ctx, g, di, passwd.MFAEventVerificationSuccess, "")
}

func AssertLastOTPEvent(ctx context.Context, g *gomega.WithT, di *LoginDI, expectEvt passwd.MFAEvent, expectUser string) passwd.OTP {
	evt := di.MFAEventRecorder.Last()
	if expectEvt != 0 {
		g.Expect(evt).ToNot(BeZero(), "last OTP event should be available")
		g.Expect(evt.Event).To(Equal(expectEvt), "last OTP event should have correct type")
		if len(expectUser) != 0 {
			acct, e := di.AccountStore.LoadAccountByUsername(ctx, expectUser)
			g.Expect(e).To(Succeed(), "load user [%s] should not fail", expectUser)
			g.Expect(evt.Principal).To(BeEquivalentTo(acct), "last OTP event should have correct principle")
		}
	} else {
		g.Expect(evt).To(BeZero(), "last OTP event should not be available")
	}
	return evt.OTP
}

