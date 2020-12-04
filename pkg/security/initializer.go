package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"go.uber.org/fx"
	"reflect"
	"sort"
	"sync"
)

/************************************
	Security Initialization
*************************************/
type initializer struct {
	initialized bool
	initializing bool
	featureConfigurers map[FeatureIdentifier]FeatureConfigurer
	configurers []Configurer
	globalAuthenticator Authenticator
}

var initializeMutex sync.Mutex

func newSecurity(globalAuth Authenticator) *initializer {
	return &initializer{
		featureConfigurers: map[FeatureIdentifier]FeatureConfigurer{},
		configurers: []Configurer{},
		globalAuthenticator: globalAuth,
	}
}

func (init *initializer) Register(configurers ...Configurer) {
	// TODO proper lock
	if err := init.validateState("register security.Configurer"); err != nil {
		panic(err)
	}
	init.configurers = append(init.configurers, configurers...)
}

func (init *initializer) RegisterFeature(featureId FeatureIdentifier, featureConfigurer FeatureConfigurer) {
	// TODO proper lock
	if err := init.validateState("register security.FeatureConfigurer"); err != nil {
		panic(err)
	}
	init.featureConfigurers[featureId] = featureConfigurer
}

func (init *initializer) validateState(action string) error {
	switch {
	case init.initialized:
		return fmt.Errorf("cannot %s: security already initialized", action)
	case init.initializing:
		return fmt.Errorf("cannot %s: security already started initializing", action)
	default:
		return nil
	}
}

func (init *initializer) Initialize(lc fx.Lifecycle, registrar *web.Registrar) error {
	initializeMutex.Lock()
	defer initializeMutex.Unlock()

	if init.initialized || init.initializing {
		return fmt.Errorf("security.Initializer.initialize cannot be called twice")
	}

	init.initializing = true

	// sort configurer
	sort.Slice(init.configurers, func(i,j int) bool {
		return order.OrderedFirstCompare(init.configurers[i], init.configurers[j])
	})

	// go through each configurer
	for _,configurer := range init.configurers {
		builder, err := init.build(configurer)
		if err != nil {
			return err
		}

		mappings := builder.Build()
		var nonMwMappings []interface{}

		// register web.MiddlewareMapping
		for _,mapping := range mappings {
			switch mapping.(type) {
			case web.MiddlewareMapping:
				mw := mapping.(web.MiddlewareMapping)
				if err := registrar.Register(mw); err != nil {
					return err
				}
				// TODO logger
				fmt.Printf("registered security middleware [%d][%s] %v -> %v \n",
					mw.Order(), mw.Name(), mw.Matcher(), reflect.ValueOf(mw.HandlerFunc()).String())
			default:
				// TODO logger
				fmt.Printf("will register security endpoints [%s]: %v\n", mapping.Name(), mapping)
				nonMwMappings = append(nonMwMappings, mapping)
			}
		}

		// register other mappings, need to be done in lifecycle
		registrar.RegisterWithLifecycle(lc, nonMwMappings...)
	}

	init.initialized = true
	init.initializing = false
	return nil
}

func (init *initializer) build(configurer Configurer) (WebSecurityMappingBuilder, error) {
	// collect security configs
	ws := newWebSecurity(NewAuthenticator(), map[string]interface{}{
		WSSharedKeyCompositeAuthSuccessHandler: NewAuthenticationSuccessHandler(),
		WSSharedKeyCompositeAuthErrorHandler: NewAuthenticationErrorHandler(),
		WSSharedKeyCompositeAccessDeniedHandler: NewAccessDeniedHandler(),
	})
	configurer.Configure(ws)

	// configure web security
	features := ws.Features()
	sort.Slice(features, func(i,j int) bool {
		return order.OrderedFirstCompare(features[i].Identifier(), features[j].Identifier())
	})

	for _, f := range features {
		fc, ok := init.featureConfigurers[f.Identifier()]
		if !ok {
			return nil, fmt.Errorf("unable to build security feature %T: no FeatureConfigurer found", f)
		}

		err := fc.Apply(f, ws)
		// mark applied
		ws.applied[f.Identifier()] = struct{}{}
		if err != nil {
			return nil, err
		}
	}

	if err := init.process(ws); err != nil {
		return nil, err
	}
	return ws, nil
}

func (init *initializer) process(ws *webSecurity) error {
	if len(ws.handlers) == 0 {
		return fmt.Errorf("no middleware were configuered for WebSecurity %v", ws)
	}

	switch {
	case !hasConcreteAuthenticator(ws.authenticator) && !hasConcreteAuthenticator(init.globalAuthenticator):
		//return fmt.Errorf("no concrete authenticator is configured for WebSecurity %v, and global authenticator is not configurered neither", ws)
		ws.authenticator.(*CompositeAuthenticator).Add(&AnonymousAuthenticator{})
	case !hasConcreteAuthenticator(ws.authenticator):
		// ws has no concrete authenticator, but global authenticator is configured, use it
		if _,ok := init.globalAuthenticator.(*CompositeAuthenticator); ok {
			ws.authenticator.(*CompositeAuthenticator).Merge(init.globalAuthenticator.(*CompositeAuthenticator))
		} else {
			ws.authenticator.(*CompositeAuthenticator).Add(init.globalAuthenticator)
		}
	}
	return nil
}

func hasConcreteAuthenticator(auth Authenticator) bool {
	if auth == nil {
		return false
	}

	composite, ok := auth.(*CompositeAuthenticator)
	return !ok || len(composite.authenticators) != 0
}






