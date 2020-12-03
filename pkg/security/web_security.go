package security

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
	"net/http"
)

/*****************************
	webSecurity Impl
******************************/
type webSecurity struct {
	routeMatcher     web.RouteMatcher
	conditionMatcher web.MWConditionMatcher
	handlers         []interface{}
	features         []Feature
	applied          map[FeatureIdentifier]struct{}
	shared           map[string]interface{}
	authenticator    Authenticator
}

func newWebSecurity(authenticator Authenticator) *webSecurity {
	return &webSecurity{
		handlers:      []interface{}{},
		features:      []Feature{},
		applied:       map[FeatureIdentifier]struct{}{},
		shared:        map[string]interface{}{},
		authenticator: authenticator,
	}
}

/* WebSecurity interface */
func (ws *webSecurity) Features() []Feature {
	return ws.features
}

func (ws *webSecurity) Route(rm web.RouteMatcher) WebSecurity {
	if ws.routeMatcher != nil {
		ws.routeMatcher = ws.routeMatcher.Or(rm)
	} else {
		ws.routeMatcher = rm
	}
	return ws
}

func (ws *webSecurity) Condition(mwcm web.MWConditionMatcher) WebSecurity {
	if ws.conditionMatcher != nil {
		ws.conditionMatcher = ws.conditionMatcher.Or(mwcm)
	} else {
		ws.conditionMatcher = mwcm
	}
	return ws
}

func (ws *webSecurity) With(f Feature) WebSecurity {
	existing := ws.Enable(f)
	if existing != f {
		panic(fmt.Errorf("cannot re-enable feature [%v] using With()", f.Identifier()))
	}
	return ws
}

func (ws *webSecurity) Add(handlers ...interface{}) WebSecurity {
	for i, h := range handlers {
		v, err := ws.toAcceptedHandler(h)
		if err != nil {
			panic(err)
		}
		handlers[i] = v
	}
	ws.handlers = append(ws.handlers, handlers...)
	return ws
}

func (ws *webSecurity) Remove(handlers ...interface{}) WebSecurity {
	for _, h := range handlers {
		v, err := ws.toAcceptedHandler(h)
		if err != nil {
			panic(err)
		}
		ws.handlers = remove(ws.handlers, v)
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
func (ws *webSecurity) Enable(f Feature) Feature {
	if _,exists := ws.applied[f.Identifier()]; exists {
		panic(fmt.Errorf("attempt to configure security feature [%v] after it has been applied", f.Identifier()))
	}

	if i := findFeatureIndex(ws.features, f); i >= 0 {
		// already have this feature
		return ws.features[i]
	}
	ws.features = append(ws.features, f)
	return f
}

func (ws *webSecurity) Disable(f Feature) {
	if i := findFeatureIndex(ws.features, f); i >= 0 {
		// already have this feature
		copy(ws.features[i:], ws.features[i + 1:])
		ws.features[len(ws.features) - 1] = nil
		ws.features = ws.features[:len(ws.features) - 1]
	}
}

/* WebSecurityReader interface */
func (ws *webSecurity) GetRoute() web.RouteMatcher {
	return ws.routeMatcher
}

func (ws *webSecurity) GetCondition() web.MWConditionMatcher {
	return ws.conditionMatcher
}

func (ws *webSecurity) GetHandlers() []interface{} {
	return ws.handlers
}

/* WebSecurityMappingBuilder interface */
func (ws *webSecurity) Build() []web.Mapping {
	mappings := make([]web.Mapping, len(ws.handlers))

	for i, v := range ws.handlers {
		var mapping web.Mapping

		if _, ok := v.(MiddlewareTemplate); ok {
			// non-interface types
			mapping = ws.buildFromMiddlewareTemplate(v.(MiddlewareTemplate))
		} else {
			// interface types
			switch v.(type) {
			case web.GenericMapping:
				mapping = v.(web.GenericMapping)
			case web.StaticMapping:
				mapping = v.(web.StaticMapping)
			case web.MvcMapping:
				mapping = v.(web.MvcMapping)
			case web.MiddlewareMapping:
				mapping = v.(web.MiddlewareMapping)
			default:
				panic(fmt.Errorf("unable to build security mappings from unsupported WebSecurity handler [%T]", v))
			}
		}
		mappings[i] = mapping
	}
	return mappings
}

// Other interfaces
func (ws *webSecurity) String() string {
	// TODO
	return fmt.Sprintf("matcher=%v condition=%v features=%v", ws.routeMatcher, ws.conditionMatcher, ws.features)
}

// unexported
func (ws *webSecurity) buildFromMiddlewareTemplate(tmpl MiddlewareTemplate) web.Mapping {
	builder := (*middleware.MappingBuilder)(tmpl)
	if ws.routeMatcher == nil {
		ws.routeMatcher = matcher.AnyRoute()
	}
	builder = builder.ApplyTo(ws.routeMatcher)

	if ws.conditionMatcher != nil {
		builder = builder.WithCondition(WebConditionFunc(ws.conditionMatcher) )
	}
	return builder.Build()
}

// toAcceptedHandler perform validation and some type casting on the interface
func (ws *webSecurity) toAcceptedHandler(v interface{}) (interface{}, error) {
		// non-interface types
		if _, ok := v.(*middleware.MappingBuilder); ok {
			return MiddlewareTemplate(v.(*middleware.MappingBuilder)), nil
		} else if _, ok := v.(MiddlewareTemplate); ok {
			return v, nil
		}

		// interface types
		switch v.(type) {
		case web.GenericMapping:
		case web.StaticMapping:
		case web.MvcMapping:
		case web.MiddlewareMapping:
		default:
			return nil, fmt.Errorf("unsupported WebSecurity handler [%T]", v)
		}
	return v, nil
}

func remove(slice []interface{}, item interface{}) []interface{} {
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

func WebConditionFunc(matcher web.MWConditionMatcher) web.MWConditionFunc {
	return func(ctx context.Context, req *http.Request) bool {
		if matches, err := matcher.MatchesWithContext(ctx, req); err != nil {
			return false
		} else {
			return matches
		}
	}
}


