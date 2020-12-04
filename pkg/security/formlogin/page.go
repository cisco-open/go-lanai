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
)

type DefaultLoginFormController struct {
	template string
	usernameParam string
	passwordParam string
	loginProcessUrl string
}

func NewDefaultLoginFormController(template string, usernameParam string, passwordParam string, loginProcessUrl string) *DefaultLoginFormController {
	return &DefaultLoginFormController{
		template: template,
		usernameParam: usernameParam,
		passwordParam: passwordParam,
		loginProcessUrl: loginProcessUrl,
	}
}

type LoginRequest struct {
	Error bool `form:"error"`
}

func (c *DefaultLoginFormController) Mappings() []web.Mapping {
	return []web.Mapping{
		template.NewBuilder().Get("/login").HandlerFunc(c.LoginForm).Build(),
	}
}

func (c *DefaultLoginFormController) LoginForm(ctx context.Context, r *LoginRequest) (*template.ModelView, error) {
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
		View: c.template,
		Model: model,
	}, nil
}
