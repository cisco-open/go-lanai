package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
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
	Cookie    *CookieProperties
}

type CookieProperties struct {
	Domain string `json:"domain"`
	MaxAge int `json:"max-age"`
	Secure bool `json:"secure"`
	HttpOnly bool `json:"http-only"`
	SameSite string `json:"same-site"`
}

//NewSessionProperties create a SessionProperties with default values
func NewSessionProperties() *SessionProperties {
	return &SessionProperties {
		Cookie: &CookieProperties{ },
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
