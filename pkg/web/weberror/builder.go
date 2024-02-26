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

package weberror

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/web"
    "github.com/cisco-open/go-lanai/pkg/web/matcher"
)

type MappingBuilder struct {
	name       string
	matcher    web.RouteMatcher
	order      int
	condition     web.RequestMatcher
	translateFunc web.ErrorTranslateFunc
}

func New(name ...string) *MappingBuilder {
	n := "anonymous"
	if len(name) != 0 {
		n = name[0]
	}
	return &MappingBuilder{
		name: n,
		matcher: matcher.AnyRoute(),
	}
}

/*****************************
	Public
******************************/

func (b *MappingBuilder) Name(name string) *MappingBuilder {
	b.name = name
	return b
}

func (b *MappingBuilder) Order(order int) *MappingBuilder {
	b.order = order
	return b
}

func (b *MappingBuilder) With(translator web.ErrorTranslator) *MappingBuilder {
	b.translateFunc = translator.Translate
	return b
}

func (b *MappingBuilder) ApplyTo(matcher web.RouteMatcher) *MappingBuilder {
	b.matcher = matcher
	return b
}

func (b *MappingBuilder) Use(translateFunc web.ErrorTranslateFunc) *MappingBuilder {
	b.translateFunc = translateFunc
	return b
}

func (b *MappingBuilder) WithCondition(condition web.RequestMatcher) *MappingBuilder {
	b.condition = condition
	return b
}

func (b *MappingBuilder) Build() web.ErrorTranslateMapping {
	if b.matcher == nil {
		b.matcher = matcher.AnyRoute()
	}
	if b.name == "" {
		b.name = fmt.Sprintf("%v", b.matcher)
	}
	if b.translateFunc == nil {
		panic(fmt.Errorf("unable to build '%s' error translation mapping: error translate function is required. please use With(...) or Use(...)", b.name))
	}
	return web.NewErrorTranslateMapping(b.name, b.order, b.matcher, b.condition, b.translateFunc)
}

/*****************************
	Helpers
******************************/




