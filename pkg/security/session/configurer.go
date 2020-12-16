package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

var (
	FeatureId = security.SimpleFeatureId("Session")
)

// We currently don't have any stuff to configure
type Feature struct {
	//TODO: allow configuring a getMaxSessions function
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
	store Store
}

func newSessionConfigurer(store Store) *Configurer {
	return &Configurer{
		store: store,
	}
}

func (sc *Configurer) Apply(_ security.Feature, ws security.WebSecurity) error {
	// configure middleware
	manager := NewManager(sc.store)

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		Order(security.MWOrderSessionHandling).
		Use(manager.SessionHandlerFunc())

	authPersist := middleware.NewBuilder("sessionMiddleware").
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

	concurrentSessionHandler := &ConcurrentSessionHandler{
		sessionStore: sc.store,
		getMaxSessions: func() int {
			//TODO: get this from configuration file
			return 5
		},
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