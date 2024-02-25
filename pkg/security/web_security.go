// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package security

import (
    "context"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/mapping"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

/*****************************
	webSecurity Impl
******************************/
type webSecurity struct {
	context          context.Context
	routeMatcher     web.RouteMatcher
	conditionMatcher web.RequestMatcher
	handlers         []interface{}
	features         []Feature
	shared           map[string]interface{}
	authenticator    Authenticator
	applied          map[FeatureIdentifier]struct{}
	featuresChanged  bool
}

func newWebSecurity(ctx context.Context, authenticator Authenticator, shared map[string]interface{}) *webSecurity {
	return &webSecurity{
		context:       ctx,
		handlers:      []interface{}{},
		features:      []Feature{},
		applied:       map[FeatureIdentifier]struct{}{},
		shared:        shared,
		authenticator: authenticator,
	}
}

/* WebSecurity interface */
func (t webSecurity) Context() context.Context {
	if t.context == nil {
		return context.TODO()
	}
	return t.context
}

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

func (ws *webSecurity) Condition(mwcm web.RequestMatcher) WebSecurity {
	if ws.conditionMatcher != nil {
		ws.conditionMatcher = ws.conditionMatcher.Or(mwcm)
	} else {
		ws.conditionMatcher = mwcm
	}
	return ws
}

func (ws *webSecurity) AndCondition(mwcm web.RequestMatcher) WebSecurity {
	if ws.conditionMatcher != nil {
		ws.conditionMatcher = ws.conditionMatcher.And(mwcm)
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
	ws.featuresChanged = true
	ws.features = append(ws.features, f)
	return f
}

func (ws *webSecurity) Disable(f Feature) {
	if i := findFeatureIndex(ws.features, f); i >= 0 {
		// already have this feature
		ws.featuresChanged = true
		copy(ws.features[i:], ws.features[i + 1:])
		ws.features[len(ws.features) - 1] = nil
		ws.features = ws.features[:len(ws.features) - 1]
	}
}

/* WebSecurityReader interface */
func (ws *webSecurity) GetRoute() web.RouteMatcher {
	return ws.routeMatcher
}

func (ws *webSecurity) GetCondition() web.RequestMatcher {
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

		switch v.(type) {
		case MiddlewareTemplate:
			mapping = ws.buildFromMiddlewareTemplate(v.(MiddlewareTemplate))
		case SimpleMappingTemplate:
			mapping = ws.buildFromSimpleMappingTemplate(v.(SimpleMappingTemplate))
		case web.SimpleMapping:
			mapping = v.(web.SimpleMapping)
		case web.StaticMapping:
			mapping = v.(web.StaticMapping)
		case web.MvcMapping:
			mapping = v.(web.MvcMapping)
		case web.MiddlewareMapping:
			mapping = v.(web.MiddlewareMapping)
		default:
			panic(fmt.Errorf("unable to build security mappings from unsupported WebSecurity handler [%T]", v))
		}
		mappings[i] = mapping
	}
	return mappings
}

// Other interfaces
func (ws *webSecurity) String() string {
	fids := make([]FeatureIdentifier, len(ws.features))
	for i, f := range ws.features {
		fids[i] = f.Identifier()
	}
	return fmt.Sprintf("matcher=%v, condition=%v, features=%v", ws.routeMatcher, ws.conditionMatcher, fids)
}

func (ws *webSecurity) GoString() string {
	return ws.String()
}

// unexported
func (ws *webSecurity) buildFromMiddlewareTemplate(tmpl MiddlewareTemplate) web.Mapping {
	builder := (*middleware.MappingBuilder)(tmpl)
	if ws.routeMatcher == nil {
		ws.routeMatcher = matcher.AnyRoute()
	}

	if builder.GetRouteMatcher() == nil {
		builder = builder.ApplyTo(ws.routeMatcher)
	}

	if ws.conditionMatcher != nil && builder.GetCondition() == nil {
		builder = builder.WithCondition(ws.conditionMatcher)
	}
	return builder.Build()
}

func (ws *webSecurity) buildFromSimpleMappingTemplate(tmpl SimpleMappingTemplate) web.Mapping {
	builder := (*mapping.MappingBuilder)(tmpl)
	if ws.routeMatcher == nil {
		ws.routeMatcher = matcher.AnyRoute()
	}

	if ws.conditionMatcher != nil && builder.GetCondition() == nil {
		builder = builder.Condition(ws.conditionMatcher)
	}
	return builder.Build()
}

// toAcceptedHandler perform validation and some type casting on the interface
func (ws *webSecurity) toAcceptedHandler(v interface{}) (interface{}, error) {
		// non-interface types
		if casted, ok := v.(*middleware.MappingBuilder); ok {
			return MiddlewareTemplate(casted), nil
		} else if casted, ok := v.(*mapping.MappingBuilder); ok {
			return SimpleMappingTemplate(casted), nil
		}

		// interface types
		switch v.(type) {
		case MiddlewareTemplate:
		case SimpleMappingTemplate:
		case web.SimpleMapping:
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



