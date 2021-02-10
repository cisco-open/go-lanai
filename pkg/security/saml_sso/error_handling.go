package saml_auth

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"errors"
	"net/http"
)

type SamlErrorHandler struct {

}

func NewSamlErrorHandler() *SamlErrorHandler {
	return &SamlErrorHandler{}
}

func (h *SamlErrorHandler) HandleError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	if errors.Is(err, security.ErrorTypeSaml) {
		code := http.StatusInternalServerError
		if translator, ok := err.(SamlSsoErrorTranslator); ok {
			code = translator.TranslateHttpStatusCode()
		}
		security.WriteErrorAsHtml(c, rw, code, err)
	}
}