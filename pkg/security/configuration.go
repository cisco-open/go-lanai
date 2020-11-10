package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"reflect"
	"sync"
)

/****************************************
	Type definitions for
    specifying web security specs
*****************************************/
type MiddlewareTemplate *middleware.MappingBuilder

type Feature interface {
	ConfigurerType() reflect.Type
}

type WebSecurity interface {
	ApplyTo(r web.RouteMatcher) WebSecurity
	Add(MiddlewareTemplate) WebSecurity
	Remove(MiddlewareTemplate) WebSecurity
	Features() []Feature
}

// FeatureModifier add or remove features. \
// Should not used directly by service
// use corresponding feature's Configure(WebSecurity) instead
type FeatureModifier interface {
	Enable(Feature) error
	Disable(Feature)
}

type WebSecurityMiddlewareBuilder interface {
	Build() []web.MiddlewareMapping
}

/*****************************
	webSecurity Impl
******************************/
type webSecurity struct {
	routeMatcher        web.RouteMatcher
	condition           web.ConditionalMiddlewareFunc
	middlewareTemplates []MiddlewareTemplate
	features            []Feature
}

func newWebSecurity() *webSecurity {
	return &webSecurity {
		middlewareTemplates: []MiddlewareTemplate {},
		features: []Feature {},
	}
}

/* WebSecurity interface */
func (ws *webSecurity) Features() []Feature {
	return ws.features
}

func (ws *webSecurity) ApplyTo(r web.RouteMatcher) WebSecurity {
	ws.routeMatcher = r
	return ws
}

func (ws *webSecurity) Add(b MiddlewareTemplate) WebSecurity {
	ws.middlewareTemplates = append(ws.middlewareTemplates, b)
	return ws
}

func (ws *webSecurity) Remove(b MiddlewareTemplate) WebSecurity {
	ws.middlewareTemplates = remove(ws.middlewareTemplates, b)
	return ws
}

/* FeatureModifier interface */
func (ws *webSecurity) Enable(f Feature) error {
	if i := findFeatureIndex(ws.features, f); i >= 0 {
		// already have this feature
		return fmt.Errorf("cannot re-enable security feature %T", f)
	}
	ws.features = append(ws.features, f)
	return nil
}

func (ws *webSecurity) Disable(f Feature) {
	if i := findFeatureIndex(ws.features, f); i >= 0 {
		// already have this feature
		copy(ws.features[i:], ws.features[i + 1:])
		ws.features[len(ws.features) - 1] = nil
		ws.features = ws.features[:len(ws.features) - 1]
	}
}

/* WebSecurityMiddlewareBuilder interface */
func (ws *webSecurity) Build() []web.MiddlewareMapping {
	mappings := make([]web.MiddlewareMapping, len(ws.middlewareTemplates))

	for i, template := range ws.middlewareTemplates {
		builder := (*middleware.MappingBuilder)(template)
		if ws.routeMatcher == nil {
			ws.routeMatcher = matcher.Any()
		}
		builder = builder.ApplyTo(ws.routeMatcher)

		if ws.condition != nil {
			builder = builder.WithCondition(ws.condition)
		}

		mappings[i] = builder.Build()
	}
	return mappings
}

func remove(slice []MiddlewareTemplate, item MiddlewareTemplate) []MiddlewareTemplate {
	for i,obj := range slice {
		if obj != item {
			continue
		}
		copy(slice[i:], slice[i + 1:])
		slice[len(slice) - 1] = nil
		return slice[:len(slice)-1]
	}
	return slice
}

func findFeatureIndex(slice []Feature, f Feature ) int {
	for i, obj := range slice {
		if f.ConfigurerType().ConvertibleTo(obj.ConfigurerType()) {
			return i
		}
	}
	return -1
}

/************************************
	Interfaces for other modules
*************************************/
type Configurer interface {
	Configure(WebSecurity)
}

// FeatureConfigurer not intended to be used directly in service
type FeatureConfigurer interface {
	Build(Feature) ([]MiddlewareTemplate, error)
}

/************************************
	Security Initialization
*************************************/
type Initializer interface {
	// Register is the entry point for all security configuration.
	// Microservice or other library packages typically call this in Provide or Invoke functions
	// Note: use this function inside fx.Lifecycle takes no effect
	Register(...Configurer)
}

type FeatureRegistrar interface {
	// RegisterFeatureConfigurer is typically used by feature packages, such as session, oauth, etc
	// not intended to be used directly in service
	RegisterFeatureConfigurer(reflect.Type, FeatureConfigurer)
}

type initializer struct {
	initialized bool
	initializing bool
	featureConfigurers map[reflect.Type]FeatureConfigurer
	configurers []Configurer
}

var initializeMutex sync.Mutex

func New() Initializer {
	return &initializer{
		featureConfigurers: map[reflect.Type]FeatureConfigurer{},
		configurers: []Configurer{},
	}
}

func (init *initializer) Register(configurers ...Configurer) {
	// TODO proper lock
	if err := init.validateState("register security.Configurer"); err != nil {
		panic(err)
	}
	init.configurers = append(init.configurers, configurers...)
}

func (init *initializer) RegisterFeatureConfigurer(configurerType reflect.Type, configurer FeatureConfigurer) {
	// TODO proper lock
	if err := init.validateState("register security.FeatureConfigurer"); err != nil {
		panic(err)
	}
	init.featureConfigurers[configurerType] = configurer
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

func (init *initializer) initialize(registrar *web.Registrar) error {
	initializeMutex.Lock()
	defer initializeMutex.Unlock()

	if init.initialized || init.initializing {
		return fmt.Errorf("security.Initializer.initialize cannot be called twice")
	}

	// TODO sort configurer
	init.initializing = true
	for _,configurer := range init.configurers {
		builder, err := init.build(configurer)
		if err != nil {
			return err
		}
		mappings := builder.Build()
		for _,mapping := range mappings {
			if err := registrar.Register(mapping); err != nil {
				return err
			}
			// TODO logger
			fmt.Printf("registered security middleware [%d][%s] %v -> %v \n",
				mapping.Order(), mapping.Name(), mapping.Matcher(), reflect.TypeOf(mapping.HandlerFunc()))
		}
	}

	init.initialized = true
	init.initializing = false
	return nil
}

func (init *initializer) build(configurer Configurer) (WebSecurityMiddlewareBuilder, error) {
	// collect security configs
	webSecurity := newWebSecurity()
	configurer.Configure(webSecurity)

	// configure web security
	templates := webSecurity.middlewareTemplates
	for _, f := range webSecurity.Features() {
		fc, ok := init.featureConfigurers[f.ConfigurerType()]
		if !ok {
			return nil, fmt.Errorf("unable to build security feature %T: no FeatureConfigurer found", f)
		}

		additional, err := fc.Build(f)
		if err != nil {
			return nil, err
		}
		templates = append(templates, additional...)
	}
	webSecurity.middlewareTemplates = templates
	return webSecurity, nil
}





