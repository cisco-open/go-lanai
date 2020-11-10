package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
	"reflect"
	"sync"
)

/************************************
	Security Initialization
*************************************/
type initializer struct {
	initialized bool
	initializing bool
	featureConfigurers map[reflect.Type]FeatureConfigurer
	configurers []Configurer
}

var initializeMutex sync.Mutex

func newSecurity() *initializer {
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

func (init *initializer) Initialize(registrar *web.Registrar) error {
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





