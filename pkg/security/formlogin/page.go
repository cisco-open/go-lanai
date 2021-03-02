package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
)

const (
	LoginModelKeyUsernameParam      = "usernameParam"
	LoginModelKeyPasswordParam      = "passwordParam"
	LoginModelKeyLoginProcessUrl    = "loginProcessUrl"
	LoginModelKeyRememberedUsername = "rememberedUsername"
	LoginModelKeyOtpParam           = "otpParam"
	LoginModelKeyMfaVerifyUrl       = "mfaVerifyUrl"
	LoginModelKeyMfaRefreshUrl      = "mfaRefreshUrl"
)

type DefaultFormLoginController struct {
	loginTemplate   string
	loginProcessUrl string
	usernameParam   string
	passwordParam   string

	mfaTemplate   string
	mfaVerifyUrl  string
	mfaRefreshUrl string
	otpParam      string
}

type PageOptionsFunc func(*DefaultFormLoginPageOptions)

type DefaultFormLoginPageOptions struct {
	LoginTemplate   string
	UsernameParam   string
	PasswordParam   string
	LoginProcessUrl string

	MfaTemplate   string
	OtpParam      string
	MfaVerifyUrl  string
	MfaRefreshUrl string
}

func NewDefaultLoginFormController(options...PageOptionsFunc) *DefaultFormLoginController {
	opts := DefaultFormLoginPageOptions{}
	for _,f := range options {
		f(&opts)
	}

	return &DefaultFormLoginController{
		loginTemplate:   opts.LoginTemplate,
		loginProcessUrl: opts.LoginProcessUrl,
		usernameParam:   opts.UsernameParam,
		passwordParam:   opts.PasswordParam,

		mfaTemplate:   opts.MfaTemplate,
		mfaVerifyUrl:  opts.MfaVerifyUrl,
		mfaRefreshUrl: opts.MfaRefreshUrl,
		otpParam:      opts.OtpParam,
	}
}

type LoginRequest struct {
	Error bool `form:"error"`
}

type OTPVerificationRequest struct {
	Error bool `form:"error"`
}

func (c *DefaultFormLoginController) Mappings() []web.Mapping {
	return []web.Mapping{
		template.New().Get("/login").HandlerFunc(c.LoginForm).Build(),
		template.New().Get("/login/mfa").HandlerFunc(c.OtpVerificationForm).Build(),
	}
}

func (c *DefaultFormLoginController) LoginForm(ctx context.Context, r *LoginRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyUsernameParam: c.usernameParam,
		LoginModelKeyPasswordParam: c.passwordParam,
		LoginModelKeyLoginProcessUrl: c.loginProcessUrl,
	}

	s := session.Get(ctx)
	if s != nil {
		if err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error); errOk && r.Error {
			model[template.ModelKeyError] = err
		}

		if username, usernameOk := s.Flash(c.usernameParam).(string); usernameOk {
			model[c.usernameParam] = username
		}
	}

	if gc := web.GinContext(ctx); gc != nil {
		if remembered, e := gc.Cookie(CookieKeyRememberedUsername); e == nil && remembered != "" {
			model[LoginModelKeyRememberedUsername] = remembered
		}
	}

	return &template.ModelView{
		View: c.loginTemplate,
		Model: model,
	}, nil
}

func (c *DefaultFormLoginController) OtpVerificationForm(ctx context.Context, r *OTPVerificationRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyOtpParam:      c.otpParam,
		LoginModelKeyMfaVerifyUrl:  c.mfaVerifyUrl,
		LoginModelKeyMfaRefreshUrl: c.mfaRefreshUrl,
	}

	s := session.Get(ctx)
	if s != nil {
		if err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error); errOk && r.Error {
			model[template.ModelKeyError] = err
		}
	}

	return &template.ModelView{
		View: c.mfaTemplate,
		Model: model,
	}, nil
}
