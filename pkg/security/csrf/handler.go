package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"github.com/gin-gonic/gin"
	"net/http"
)

type ChangeCsrfHanlder struct{
	csrfTokenStore TokenStore
}

func (h *ChangeCsrfHanlder) HandleAuthenticationSuccess(c context.Context, _ *http.Request, _ http.ResponseWriter, from, to security.Authentication) {
	if !security.IsBeingAuthenticated(from, to) {
		return
	}

	if gc, ok := c.(*gin.Context); ok {
		t, err := h.csrfTokenStore.LoadToken(gc)

		if err != nil {
			panic(security.NewInternalError(err.Error()))
		}

		if t != nil {
			t = h.csrfTokenStore.Generate(gc, t.ParameterName, t.HeaderName)
			h.csrfTokenStore.SaveToken(gc, t)
			gc.Set(web.ContextKeyCsrf, t)
		}
	}
}