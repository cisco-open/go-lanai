package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/passwd"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"encoding/gob"
	"go.uber.org/fx"
	"net/http"
	"path"
	"strings"
	"time"
)

var SessionModule = &bootstrap.Module{
	Name: "session",
	Precedence: security.MinSecurityPrecedence + 10,
	Options: []fx.Option{
		fx.Provide(security.BindSessionProperties, newSessionStore, newRequestCacheMatcher),
		fx.Invoke(register),
	},
}


func init() {
	bootstrap.Register(SessionModule)

	GobRegister()
	security.GobRegister()
	passwd.GobRegister()
}

func GobRegister() {
	gob.Register([]interface{}{})
}


func register(init security.Registrar, sessionStore Store) {
	configurer := newSessionConfigurer(sessionStore)
	init.(security.FeatureRegistrar).RegisterFeature(FeatureId, configurer)
}

func newSessionStore(sessionProps security.SessionProperties, serverProps web.ServerProperties, connection *redis.Connection) Store {
	// configure session store
	var sameSite http.SameSite
	switch strings.ToLower(sessionProps.Cookie.SameSite) {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}

	idleTimeout, err := time.ParseDuration(sessionProps.IdleTimeout)
	if err != nil {
		panic(err)
	}
	absTimeout, err := time.ParseDuration(sessionProps.AbsoluteTimeout)
	if err != nil {
		panic(err)
	}

	configureOptions := func(options *Options) {
		options.Path = path.Clean("/" + serverProps.ContextPath)
		options.Domain = sessionProps.Cookie.Domain
		options.MaxAge = sessionProps.Cookie.MaxAge
		options.Secure = sessionProps.Cookie.Secure
		options.HttpOnly = sessionProps.Cookie.HttpOnly
		options.SameSite = sameSite
		options.IdleTimeout = idleTimeout
		options.AbsoluteTimeout = absTimeout
	}
	sessionStore := NewRedisStore(connection, configureOptions)

	return sessionStore
}

func newRequestCacheMatcher(sessionStore Store) web.RequestCacheMatcher {
	m := &RequestCacheMatcher{
		store: sessionStore,
	}
	return m
}