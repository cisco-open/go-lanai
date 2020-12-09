package formlogin

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/redirect"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
)

const (
	LoginModelKeyUsernameParam = "usernameParam"
	LoginModelKeyPasswordParam = "passwordParam"
	LoginModelKeyLoginProcessUrl = "loginProcessUrl"
	LoginModelKeyRememberedUsername = "rememberedUsername"
	LoginModelKeyShowError = "showError"
	LoginModelKeyOtpParam         = "otpParam"
	LoginModelKeyOtpVerifyUrl    = "otpVerifyUrl"
)

type DefaultFormLoginController struct {
	loginTemplate   string
	usernameParam   string
	passwordParam   string
	loginProcessUrl string

	mfaTemplate  string
	otpParam     string
	mfaVerifyUrl string
}

type PageOptionsFunc func(*DefaultFormLoginPageOptions)

type DefaultFormLoginPageOptions struct {
	LoginTemplate string
	UsernameParam   string
	PasswordParam   string
	LoginProcessUrl string

	MfaTemplate string
	OtpParam string
	MfaVerifyUrl string
}

func NewDefaultLoginFormController(options...PageOptionsFunc) *DefaultFormLoginController {
	opts := DefaultFormLoginPageOptions{}
	for _,f := range options {
		f(&opts)
	}

	return &DefaultFormLoginController{
		loginTemplate:   opts.LoginTemplate,
		usernameParam:   opts.UsernameParam,
		passwordParam:   opts.PasswordParam,
		loginProcessUrl: opts.LoginProcessUrl,

		mfaTemplate:  opts.MfaTemplate,
		otpParam:     opts.OtpParam,
		mfaVerifyUrl: opts.MfaVerifyUrl,
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
		template.NewBuilder().Get("/login").HandlerFunc(c.LoginForm).Build(),
		template.NewBuilder().Get("/login/otp").HandlerFunc(c.OtpVerificationForm).Build(),
	}
}

func (c *DefaultFormLoginController) LoginForm(ctx context.Context, r *LoginRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyUsernameParam: c.usernameParam,
		LoginModelKeyPasswordParam: c.passwordParam,
		LoginModelKeyLoginProcessUrl: c.loginProcessUrl,
	}

	if r.Error {
		model[LoginModelKeyShowError] = true
	}

	s := session.Get(ctx)
	if s != nil {
		if err, errOk := s.Flash(redirect.FlashKeyPreviousError).(error); errOk {
			model[template.ModelKeyError] = err
		}

		if username, usernameOk := s.Flash(c.usernameParam).(string); usernameOk {
			model[c.usernameParam] = username
		} else if username, usernameOk := s.Get(SessionKeyRememberedUsername).(string); usernameOk {
			model[LoginModelKeyRememberedUsername] = username
		}
	}

	return &template.ModelView{
		View: c.loginTemplate,
		Model: model,
	}, nil
}

func (c *DefaultFormLoginController) OtpVerificationForm(ctx context.Context, r *OTPVerificationRequest) (*template.ModelView, error) {
	model := template.Model{
		LoginModelKeyOtpParam:     c.otpParam,
		LoginModelKeyOtpVerifyUrl: c.mfaVerifyUrl,
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
