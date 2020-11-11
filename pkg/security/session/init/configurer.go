package init

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/store"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

const (
	MWOrderSessionHandling = security.HighestMiddlewareOrder + 100
	MWOrderAuthPersistence = MWOrderSessionHandling + 10
	SessionFeatureId       = "Session"
)

// We currently don't have any stuff to configure
type SessionFeature struct {

}

func (f *SessionFeature) Identifier() security.FeatureIdentifier {
	return SessionFeatureId
}

func Configure(ws security.WebSecurity) *SessionFeature {
	feature := &SessionFeature{}
	if fc, ok := ws.(security.FeatureModifier); ok {
		_ = fc.Enable(feature) // we ignore error here
		return feature
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

type SessionConfigurer struct {
	sessionProps security.SessionProperties
	serverProps web.ServerProperties
}

func newSessionConfigurer(sessionProps security.SessionProperties, serverProps web.ServerProperties) *SessionConfigurer {
	return &SessionConfigurer{
		sessionProps: sessionProps,
		serverProps: serverProps,
	}
}

func (sc *SessionConfigurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	// TODO
	sessionStore := newSessionStore(sc.sessionProps)
	manager := session.NewManager(sessionStore, sc.sessionProps, sc.serverProps)

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		Order(MWOrderSessionHandling).
		Use(manager.SessionHandlerFunc())

	authPersist := middleware.NewBuilder("sessionMiddleware").
		Order(MWOrderAuthPersistence).
		Use(manager.AuthenticationPersistenceHandlerFunc())

	test := middleware.NewBuilder("post-sessionMiddleware").
		Order(MWOrderAuthPersistence + 10).
		Use(session.SessionDebugHandlerFunc())

	ws.Add(sessionHandler, authPersist, test)
	return nil
}

func newSessionStore(properties security.SessionProperties) session.Store {
	secret := []byte(properties.Secret)
	switch properties.StoreType {
	case security.SessionStoreTypeMemory:
		return store.NewMemoryStore(secret)
	default:
		panic(fmt.Errorf("unsupported session storage"))
	}
}