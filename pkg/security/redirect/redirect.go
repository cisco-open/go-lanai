package redirect

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
	urlutils "net/url"
	"path"
)

const (
	FlashKeyPreviousError      = "error"
	FlashKeyPreviousStatusCode = "status"
)

// RedirectHandler implements multiple interface for authentication and error handling strategies
//goland:noinspection GoNameStartsWithPackageName
type RedirectHandler struct {
	sc       int
	location string
}

func NewRedirectWithRelativePath(path string) *RedirectHandler {
	url, err := urlutils.Parse(path)
	if err != nil {
		panic(err)
	}

	url.ForceQuery = true
	return &RedirectHandler{
		sc:       302,
		location: path,
	}
}

func NewRedirectWithURL(urlStr string) *RedirectHandler {
	url, err := urlutils.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	url.ForceQuery = true
	return &RedirectHandler{
		sc:       302,
		location: urlStr,
	}
}

// security.AuthenticationEntryPoint
func (ep *RedirectHandler) Commence(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError: err,
		FlashKeyPreviousStatusCode: http.StatusUnauthorized,
	})
}

// security.AccessDeniedHandler
func (ep *RedirectHandler) HandleAccessDenied(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError: err,
		FlashKeyPreviousStatusCode: http.StatusForbidden,
	})
}

// security.AuthenticationSuccessHandler
func (ep *RedirectHandler) HandleAuthenticationSuccess(c context.Context, r *http.Request, rw http.ResponseWriter, auth security.Authentication) {
	ep.doRedirect(c, r, rw, nil)
}

// security.AuthenticationErrorHandler
func (ep *RedirectHandler) HandleAuthenticationError(c context.Context, r *http.Request, rw http.ResponseWriter, err error) {
	ep.doRedirect(c, r, rw, map[string]interface{}{
		FlashKeyPreviousError: err,
		FlashKeyPreviousStatusCode: http.StatusUnauthorized,
	})
}

func (ep *RedirectHandler) doRedirect(c context.Context, r *http.Request, rw http.ResponseWriter, flashes map[string]interface{}) {
	// save flashes
	if flashes != nil && len(flashes) != 0 {
		s := session.Get(c)
		for k,v := range flashes {
			s.AddFlash(v, k)
		}
	}

	location,_ := urlutils.ParseRequestURI(ep.location)
	if !location.IsAbs() {
		// relative path was used, try to add context path
		contextPath, ok := c.Value(web.ContextKeyContextPath).(string)
		if ok {
			location.Path = path.Join(contextPath, location.Path)
		}
	}

	// redirect
	http.Redirect(rw, r, location.RequestURI(), ep.sc)
}

