package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/route"
	"fmt"
	"net/http"
)

/*****************************
	webSecurity Impl
******************************/
type webSecurity struct {
	routeMatcher        web.RouteMatcher
	conditionMatcher    web.MWConditionMatcher
	middlewareTemplates []MiddlewareTemplate
	features            []Feature
	shared 				map[string]interface{}
	authenticator 		Authenticator
}

func newWebSecurity(authenticator Authenticator) *webSecurity {
	return &webSecurity{
		middlewareTemplates: []MiddlewareTemplate{},
		features: []Feature{},
		shared: map[string]interface{}{},
		authenticator: authenticator,
	}
}

/* WebSecurity interface */
func (ws *webSecurity) Features() []Feature {
	return ws.features
}

func (ws *webSecurity) Route(rm web.RouteMatcher) WebSecurity {
	ws.routeMatcher = rm
	return ws
}

func (ws *webSecurity) Condition(mwcm web.MWConditionMatcher) WebSecurity {
	ws.conditionMatcher = mwcm
	return ws
}

func (ws *webSecurity) With(f Feature) WebSecurity {
	if err := ws.Enable(f); err != nil {
		panic(err)
	}
	return ws
}

func (ws *webSecurity) Add(templates ...MiddlewareTemplate) WebSecurity {
	ws.middlewareTemplates = append(ws.middlewareTemplates, templates...)
	return ws
}

func (ws *webSecurity) Remove(templates ...MiddlewareTemplate) WebSecurity {
	for _, t := range templates {
		ws.middlewareTemplates = remove(ws.middlewareTemplates, t)
	}
	return ws
}

func (ws *webSecurity) Shared(key string) interface{} {
	return ws.shared[key]
}

func (ws *webSecurity) AddShared(key string, value interface{}) error {
	if _, exists := ws.shared[key]; exists {
		return fmt.Errorf("cannot add shared value to WebSecurity %v: key [%s] already exists", ws, key)
	}
	ws.shared[key] = value
	return nil
}

func (ws *webSecurity) Authenticator() Authenticator {
	return ws.authenticator
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
			ws.routeMatcher = route.AnyRoute()
		}
		builder = builder.ApplyTo(ws.routeMatcher)

		if ws.conditionMatcher != nil {
			builder = builder.WithCondition(conditionFunc(ws.conditionMatcher) )
		}

		mappings[i] = builder.Build()
	}
	return mappings
}

// Other interfaces
func (ws *webSecurity) String() string {
	// TODO
	return fmt.Sprintf("matcher=%v condition=%v features=%v", ws.routeMatcher, ws.conditionMatcher, ws.features)
}

// unexported
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

func findFeatureIndex(slice []Feature, f Feature) int {
	for i, obj := range slice {
		if f.Identifier() == obj.Identifier() {
			return i
		}
	}
	return -1
}

func conditionFunc(matcher web.MWConditionMatcher) web.MWConditionFunc {
	return func(req *http.Request) bool {
		if matches, err := matcher.Matches(req); err != nil {
			return false
		} else {
			return matches
		}
	}
}


