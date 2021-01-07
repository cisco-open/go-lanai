package session

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/redis"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"path"
	"time"
)

var (
	FeatureId = security.SimpleFeatureId("Session")
)

// We currently don't have any stuff to configure
type Feature struct {
	maxSessionsFunc GetMaximumSessions
	requestCacheEnabled bool
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *Feature) MaxSessionFunc(maxSessionFunc GetMaximumSessions) *Feature {
	f.maxSessionsFunc = maxSessionFunc
	return f
}

//this enables request cache request preprocessor for the entire application
func (f *Feature) EnableRequestCachePreProcessor() *Feature {
	f.requestCacheEnabled = true
	return f
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
	sessionProps security.SessionProperties
	serverProps web.ServerProperties
	redisClient redis.Client

	//cached store instance
	store Store
	//cached request processor instance
	requestPreProcessor *CachedRequestPreProcessor
}

func newSessionConfigurer(sessionProps security.SessionProperties, serverProps web.ServerProperties, redisClient redis.Client) *Configurer {
	return &Configurer{
		sessionProps: sessionProps,
		serverProps: serverProps,
		redisClient: redisClient,
	}
}

func (sc *Configurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	// configure session store
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
		options.SameSite = sc.sessionProps.Cookie.SameSite()
		options.IdleTimeout = idleTimeout
		options.AbsoluteTimeout = absTimeout
	}

	if sc.store == nil {
		sc.store = NewRedisStore(sc.redisClient, configureOptions)
	}

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

	maxSessionsFunc := f.maxSessionsFunc
	if maxSessionsFunc == nil {
		maxSessions := sc.sessionProps.MaxConcurrentSession
		maxSessionsFunc = func() int {
			return maxSessions
		}
	}

	concurrentSessionHandler := &ConcurrentSessionHandler{
		sessionStore:   sc.store,
		getMaxSessions: maxSessionsFunc,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(concurrentSessionHandler)

	deleteSessionHandler := &DeleteSessionOnLogoutHandler{
		sessionStore: sc.store,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(deleteSessionHandler)

	if f.requestCacheEnabled && sc.requestPreProcessor == nil {
		sc.requestPreProcessor = &CachedRequestPreProcessor{
			store: sc.store,
		}
	}
	return nil
}

func (sc *Configurer) ProvidePreProcessor() web.RequestPreProcessor {
	if sc.requestPreProcessor == nil {
		return nil
	}
	return sc.requestPreProcessor
}