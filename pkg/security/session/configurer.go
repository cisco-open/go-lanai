package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
)

const (
	FeatureId              = "RequestSession"
)

// We currently don't have any stuff to configure
type SessionFeature struct {

}

func (f *SessionFeature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *SessionFeature {
	feature := &SessionFeature{}
	if fc, ok := ws.(security.FeatureModifier); ok {
		_ = fc.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *SessionFeature {
	return &SessionFeature{}
}

type SessionConfigurer struct {
	sessionProps security.SessionProperties
	serverProps web.ServerProperties
	connection *redis.Connection
}

func newSessionConfigurer(sessionProps security.SessionProperties, serverProps web.ServerProperties, connection *redis.Connection) *SessionConfigurer {
	return &SessionConfigurer{
		sessionProps: sessionProps,
		serverProps: serverProps,
		connection: connection,
	}
}

func (sc *SessionConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	var sameSite http.SameSite
	switch strings.ToLower(sc.sessionProps.Cookie.SameSite) {
	case "lax":
		sameSite = http.SameSiteLaxMode
	case "strict":
		sameSite = http.SameSiteStrictMode
	case "none":
		sameSite = http.SameSiteNoneMode
	default:
		sameSite = http.SameSiteDefaultMode
	}

	idleTimeout, err := time.ParseDuration(sc.sessionProps.IdleTimeout)
	if err != nil {
		return err
	}
	absTimeout, err := time.ParseDuration(sc.sessionProps.AbsoluteTimeout)
	if err != nil {
		return err
	}

	configureOptions := func(options *Options) {
		options.Path = path.Clean("/" + sc.serverProps.ContextPath)
		options.Domain = sc.sessionProps.Cookie.Domain
		options.MaxAge = sc.sessionProps.Cookie.MaxAge
		options.Secure = sc.sessionProps.Cookie.Secure
		options.HttpOnly = sc.sessionProps.Cookie.HttpOnly
		options.SameSite = sameSite
		options.IdleTimeout = idleTimeout
		options.AbsoluteTimeout = absTimeout
	}
	sessionStore := NewRedisStore(sc.connection, configureOptions)

	manager := NewManager(sessionStore)

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		Order(security.MWOrderSessionHandling).
		Use(manager.SessionHandlerFunc())

	authPersist := middleware.NewBuilder("sessionMiddleware").
		Order(security.MWOrderAuthPersistence).
		Use(manager.AuthenticationPersistenceHandlerFunc())

	test := middleware.NewBuilder("post-sessionMiddleware").
		Order(security.MWOrderAuthPersistence + 10).
		Use(SessionDebugHandlerFunc())

	ws.Add(sessionHandler, authPersist, test)
	return nil
}