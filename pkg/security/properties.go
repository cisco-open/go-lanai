package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

/***********************
	Session
************************/
const (
	SessionStoreTypeMemory = iota
	SessionStoreTypeRedis
)

type SessionStoreType int

const SessionPropertiesPrefix = "security.session"

type SessionProperties struct {
	Cookie    CookieProperties
	IdleTimeout string `json:"idle-timeout"`
	AbsoluteTimeout string `json:"absolute-timeout"`
	MaxConcurrentSession int `json:"max-concurrent-sessions"`
}

type CookieProperties struct {
	Domain string `json:"domain"`
	MaxAge int `json:"max-age"`
	Secure bool `json:"secure"`
	HttpOnly bool `json:"http-only"`
	SameSiteString string `json:"same-site"`
}

func (cp CookieProperties) SameSite() http.SameSite {
	switch strings.ToLower(cp.SameSiteString) {
	case "lax":
		return http.SameSiteLaxMode
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteDefaultMode
	}
}

//NewSessionProperties create a SessionProperties with default values
func NewSessionProperties() *SessionProperties {
	return &SessionProperties {
		Cookie: CookieProperties{
			HttpOnly: true,
		},
		IdleTimeout: "900s",
		AbsoluteTimeout: "1800s",
		MaxConcurrentSession: 0, //unlimited
	}
}

//BindSessionProperties create and bind SessionProperties, with a optional prefix
func BindSessionProperties(ctx *bootstrap.ApplicationContext) SessionProperties {
	props := NewSessionProperties()
	if err := ctx.Config().Bind(props, SessionPropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind SessionProperties"))
	}
	return *props
}
