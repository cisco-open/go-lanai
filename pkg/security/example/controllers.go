package example

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/formlogin"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
)

func NewLoginFormController() web.Controller {
	return formlogin.NewDefaultLoginFormController(func(opts *formlogin.DefaultFormLoginPageOptions) {
		opts.LoginTemplate = "login.tmpl"
		opts.LoginProcessUrl = "/login"
		opts.UsernameParam = "username"
		opts.PasswordParam = "password"
		opts.MfaTemplate = "login.tmpl"
		opts.MfaVerifyUrl = "/login/mfa"
		opts.MfaRefreshUrl = "/login/mfa/refresh"
		opts.OtpParam = "otp"
	})
}
