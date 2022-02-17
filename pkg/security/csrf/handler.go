package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
)

type ChangeCsrfHanlder struct{
	csrfTokenStore TokenStore
}

func (h *ChangeCsrfHanlder) HandleAuthenticationSuccess(c context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	// TODO: review error handling of this block
	if mc, ok := c.(utils.MutableContext); ok {
		t, err := h.csrfTokenStore.LoadToken(c)

		if err != nil {
			panic(security.NewInternalError(err.Error()))
		}

		if t != nil {
			t = h.csrfTokenStore.Generate(c, t.ParameterName, t.HeaderName)
			if e := h.csrfTokenStore.SaveToken(c, t); e != nil {
				panic(security.NewInternalError(err.Error()))
			}
			mc.Set(web.ContextKeyCsrf, t)
		}
	}
}