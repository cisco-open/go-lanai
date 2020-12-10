package csrf

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"net/http"
)

type ChangeCsrfHanlder struct{}

//TODO
func (h *ChangeCsrfHanlder) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, from, to security.Authentication) {

}