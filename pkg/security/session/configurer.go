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

func (f *Feature) EnableRequestCache(enable bool) *Feature {
	f.requestCacheEnabled = enable
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
	requestPreProcessor web.RequestPreProcessor
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

	if sc.store == nil {
		sc.store = newSessionStore(sc.sessionProps, sc.serverProps, sc.redisClient)
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

func (sc *Configurer) GetPreProcessor() web.RequestPreProcessor {
	return sc.requestPreProcessor
}

func newSessionStore(sessionProps security.SessionProperties, serverProps web.ServerProperties, client redis.Client) Store {
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
	sessionStore := NewRedisStore(client, configureOptions)

	return sessionStore
}