package security

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/matcher"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web/middleware"
	"fmt"
)

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
	return &webSecurity{
		middlewareTemplates: []MiddlewareTemplate{},
		features: []Feature{},
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

func findFeatureIndex(slice []Feature, f Feature) int {
	for i, obj := range slice {
		if f.ConfigurerType().ConvertibleTo(obj.ConfigurerType()) {
			return i
		}
	}
	return -1
}


