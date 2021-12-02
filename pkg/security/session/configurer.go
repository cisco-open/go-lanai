package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

var (
	FeatureId = security.FeatureId("Session", security.FeatureOrderSession)
)

// We currently don't have any stuff to configure
type Feature struct {
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

// Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *Feature {
	feature := &Feature{}
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *Feature {
	return &Feature{}
}

type Configurer struct {
	store                 Store
	sessionProps          security.SessionProperties
	sessionSettingService SettingService
}

func newSessionConfigurer(sessionProps security.SessionProperties, sessionStore Store, sessionSettingService SettingService) *Configurer {
	return &Configurer{
		store:           sessionStore,
		sessionProps:    sessionProps,
		sessionSettingService: sessionSettingService,
	}
}

func (sc *Configurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	// the ws shared store is to share this store with other feature configurer can have access to store.
	if ws.Shared(security.WSSharedKeySessionStore) == nil {
		_ = ws.AddShared(security.WSSharedKeySessionStore, sc.store)
	}

	// configure middleware
	manager := NewManager(sc.store)

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		Order(security.MWOrderSessionHandling).
		Use(manager.SessionHandlerFunc())

	authPersist := middleware.NewBuilder("authPersistMiddleware").
		Order(security.MWOrderAuthPersistence).
		Use(manager.AuthenticationPersistenceHandlerFunc())

	//test := middleware.NewBuilder("post-sessionMiddleware").
	//	Order(security.MWOrderAuthPersistence + 10).
	//	Use(SessionDebugHandlerFunc())

	ws.Add(sessionHandler, authPersist)

	// configure auth success/error handler
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(&ChangeSessionHandler{})
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(&DebugAuthSuccessHandler{})
	ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler).
		Add(&DebugAuthErrorHandler{})

	var settingService SettingService
	if sc.sessionSettingService == nil {
		settingService = NewDefaultSettingService(sc.sessionProps)
	} else {
		settingService = sc.sessionSettingService
	}

	concurrentSessionHandler := &ConcurrentSessionHandler{
		sessionStore:   sc.store,
		sessionSettingService: settingService,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(concurrentSessionHandler)

	deleteSessionHandler := &DeleteSessionOnLogoutHandler{
		sessionStore: sc.store,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(deleteSessionHandler)
	return nil
}