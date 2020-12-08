package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

type ChangeCsrfHanlder struct{}

func (h *ChangeCsrfHanlder) HandleAuthenticationSuccess(c context.Context, r *http.Request, w http.ResponseWriter, a security.Authentication) {

}
