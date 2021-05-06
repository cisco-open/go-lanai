package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
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

// Register is not threadsafe, usually called in "fx.Invoke" or "fx.Provide"
func (init *initializer) Register(configurers ...Configurer) {
	if err := init.validateState("register security.Configurer"); err != nil {
		panic(err)
	}
	init.configurers = append(init.configurers, configurers...)
}

// RegisterFeature is not threadsafe, usually called in "fx.Invoke" or "fx.Provide"
func (init *initializer) RegisterFeature(featureId FeatureIdentifier, featureConfigurer FeatureConfigurer) {
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

func (init *initializer) Initialize(ctx context.Context, _ fx.Lifecycle, registrar *web.Registrar) error {
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

	mergedRequestPreProcessors := make(map[web.RequestPreProcessorName]web.RequestPreProcessor)

	// go through each configurer
	for _,configurer := range init.configurers {
		builder, requestPreProcessors, err := init.build(ctx, configurer)
		if err != nil {
			return err
		}

		for k, v := range requestPreProcessors {
			if _, ok := mergedRequestPreProcessors[k]; !ok {
				mergedRequestPreProcessors[k] = v
			}
		}

		mappings := builder.Build()
		// register web.Mapping
		for _,mapping := range mappings {
			if err := registrar.Register(mapping); err != nil {
				return err
			}
			// Do some logging
			logMapping(ctx, mapping)
		}
	}

	for _, v := range mergedRequestPreProcessors {
		registrar.MustRegister(v)
	}

	init.initialized = true
	init.initializing = false
	return nil
}

func (init *initializer) build(ctx context.Context, configurer Configurer) (WebSecurityMappingBuilder, map[web.RequestPreProcessorName]web.RequestPreProcessor, error) {
	// collect security configs
	ws := newWebSecurity(ctx, NewAuthenticator(), map[string]interface{}{
		WSSharedKeyCompositeAuthSuccessHandler: NewAuthenticationSuccessHandler(),
		WSSharedKeyCompositeAuthErrorHandler: NewAuthenticationErrorHandler(),
		WSSharedKeyCompositeAccessDeniedHandler: NewAccessDeniedHandler(),
	})
	configurer.Configure(ws)

	// configure web security
	// Note: We want to allow feature's configurer to add/remove other features while in the iteration.
	//		 Adding/removing features that have lower order than the current feature should panic
	// 		 Doing so would result in performance reduction on iteration. But it's small price we are willing to pay
	sortFeatures(ws.Features())
	for i := 0; i < len(ws.Features()); i++ {
		f := ws.Features()[i]

		// get corresponding feature configurer
		fc, ok := init.featureConfigurers[f.Identifier()]
		if !ok {
			return nil, nil, fmt.Errorf("unable to build security feature %T: no FeatureConfigurer found", f)
		}

		// mark/reset some flags
		ws.applied[f.Identifier()] = struct{}{}
		ws.featuresChanged = false

		// apply configurer
		if err := fc.Apply(f, ws); err != nil {
			return nil, nil, fmt.Errorf("Error during process WebSecurity [%v]: %v", ws, err)
		}

		// the applied configurer may have enabled more features. (ws.Features() is different)
		if !ws.featuresChanged {
			continue
		}

		// handle feature change
		if err := init.handleFeaturesChanged(i, f, ws.Features()); err != nil {
			return nil, nil, fmt.Errorf("Error during process WebSecurity [%v]: %v", ws, err)
		}
	}

	if err := init.process(ws); err != nil {
		return nil, nil, err
	}

	var processors map[web.RequestPreProcessorName]web.RequestPreProcessor = nil
	if ws.Shared(WSSharedKeyRequestPreProcessors) != nil {
		processors = ws.Shared(WSSharedKeyRequestPreProcessors).(map[web.RequestPreProcessorName]web.RequestPreProcessor)
	}

	return ws, processors, nil
}

// handleFeaturesChanged is invoked if feature list changed during iteration.
// we need to
// 	1. check if current feature's index didn't change (in case elements before current were removed)
// 	2. re-sort the remaining (un-processed) features from current index
// 	3. check if any remaining features (likely newly added) has lower order than current
func (init *initializer) handleFeaturesChanged(i int, f Feature, features []Feature) error {
	if i >= len(features) - 1 {
		// last one, nothing is needed
		return nil
	}

	// step 1
	if features[i] != f {
		return fmt.Errorf("feature configurer for [%v] attempted to disable already applied features", f.Identifier())
	}

	// step 2
	sortFeatures(features[i+1:])

	// step 3
	next := features[i+1]
	if featureOrderLess(next, f) {
		return fmt.Errorf("feature configurer for [%v] attempted to enable feature [%v] which won't be processed", f.Identifier(), next.Identifier())
	}

	return nil
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

func sortFeatures(features []Feature) {
	sort.Slice(features, func(i,j int) bool {
		return featureOrderLess(features[i], features[j])
	})
}

func featureOrderLess(l Feature, r Feature) bool {
	return order.OrderedFirstCompare(l.Identifier(), r.Identifier())
}

func logMapping(ctx context.Context, mapping web.Mapping) {
	switch mapping.(type) {
	case web.MiddlewareMapping:
		mw := mapping.(web.MiddlewareMapping)
		logger.WithContext(ctx).Infof("registered security middleware [%d] [%s] %s -> %v",
			mw.Order(), mw.Name(), log.Capped(mw.Matcher(), 80), reflect.ValueOf(mw.HandlerFunc()).String())
	case web.MvcMapping:
		m := mapping.(web.MvcMapping)
		logger.WithContext(ctx).Infof("registered security MVC mapping [%s %s] [%s] %s -> endpoint",
			m.Method(), m.Path(), m.Name(), log.Capped(m.Condition(), 80))
	case web.SimpleMapping:
		m := mapping.(web.SimpleMapping)
		logger.WithContext(ctx).Infof("registered security simple mapping [%s %s] [%s] %s -> %v",
			m.Method(), m.Path(), m.Name(), log.Capped(m.Condition(), 80), reflect.ValueOf(m.HandlerFunc()).String())
	default:
		logger.WithContext(ctx).Infof("registered security mapping [%s]: %v", mapping.Name(), mapping)
	}
}



