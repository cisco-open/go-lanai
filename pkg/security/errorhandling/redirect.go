package errorhandling

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
	urlutils "net/url"
)

const (
	FlashKeyPreviousError      = "error"
	FlashKeyPreviousStatusCode = "status"
)

type RedirectErrorHandler struct {
	StatusCode int
	Location *urlutils.URL

}

func NewRedirectWithRelativePath(path string) *RedirectErrorHandler {
	url, err := urlutils.Parse(path)
	if err != nil {
		panic(err)
	}

	return &RedirectErrorHandler{
		StatusCode: 302,
		Location: url,
	}
}

func NewRedirectWithURL(urlStr string) *RedirectErrorHandler {
	url, err := urlutils.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	return &RedirectErrorHandler{
		StatusCode: 302,
		Location: url,
	}
}

// security.AuthenticationEntryPoint
func (ep *RedirectErrorHandler) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, http.StatusUnauthorized, err)
}

// security.AccessDeniedHandler
func (ep *RedirectErrorHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, http.StatusForbidden, err)
}

func (ep *RedirectErrorHandler) doRedirect(c context.Context, r *http.Request, rw http.ResponseWriter, errorSc int, err error) {
	// save error as flash
	s := session.Get(c)
	s.AddFlash(err, FlashKeyPreviousError)
	s.AddFlash(errorSc, FlashKeyPreviousStatusCode)

	loc := ep.Location.EscapedPath()
	if !ep.Location.IsAbs() {
		// relative path was used, try to add context path
		contextPath, ok := c.Value(web.ContextKeyContextPath).(string)
		if ok {
			loc = contextPath + loc
		}
	}

	// redirect
	http.Redirect(rw, r, loc, ep.StatusCode)
}

