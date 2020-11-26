package errorhandling

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/template"
	"net/http"
	urlutils "net/url"
)

type RedirectAuthenticationEntryPoint struct {
	StatusCode int
	Location *urlutils.URL

}

func NewRedirectWithRelativePath(path string) *RedirectAuthenticationEntryPoint {
	url, err := urlutils.Parse(path)
	if err != nil {
		panic(err)
	}

	return &RedirectAuthenticationEntryPoint{
		StatusCode: 302,
		Location: url,
	}
}

func NewRedirectWithURL(urlStr string) *RedirectAuthenticationEntryPoint {
	url, err := urlutils.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	return &RedirectAuthenticationEntryPoint{
		StatusCode: 302,
		Location: url,
	}
}

func (ep *RedirectAuthenticationEntryPoint) Commence(c context.Context, r *http.Request, resp http.ResponseWriter, err error) {
	// save error as flash
	s := session.Get(c)
	s.Set(template.ModelKeyError, err)
	s.Set(template.ModelKeyMessage, err.Error())

	loc := ep.Location.EscapedPath()
	if !ep.Location.IsAbs() {
		// relative path was used, try to add context path
		contextPath, ok := c.Value(web.ContextKeyContextPath).(string)
		if ok {
			loc = contextPath + loc
		}
	}

	// redirect
	http.Redirect(resp, r, loc, ep.StatusCode)
}

