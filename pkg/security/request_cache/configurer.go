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

package request_cache

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security/session/common"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"fmt"
)

var (
	FeatureId = security.FeatureId("request_cache", security.FeatureOrderRequestCache)
)

type Feature struct {
	sessionName string
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *Feature) SessionName(sessionName string) *Feature {
	f.sessionName = sessionName
	return f
}

// Configure Standard security.Feature entrypoint
func Configure(ws security.WebSecurity) *Feature {
	feature := New()
	if fc, ok := ws.(security.FeatureModifier); ok {
		return fc.Enable(feature).(*Feature)
	}
	panic(fmt.Errorf("unable to configure session: provided WebSecurity [%T] doesn't support FeatureModifier", ws))
}

// New Standard security.Feature entrypoint, DSL style. Used with security.WebSecurity
func New() *Feature {
	return &Feature{
		sessionName: common.DefaultName,
	}
}

type Configurer struct {
	//cached request preprocessor
	cachedRequestPreProcessor *CachedRequestPreProcessor
}

func newConfigurer() *Configurer {
	return &Configurer{}
}

func (sc *Configurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	if len(f.sessionName) == 0 {
		f.sessionName = common.DefaultName
	}

	if sc.cachedRequestPreProcessor == nil {
		if store, ok := ws.Shared(security.WSSharedKeySessionStore).(session.Store); ok {
			p := newCachedRequestPreProcessor(f.sessionName, store)
			sc.cachedRequestPreProcessor = p

			if ws.Shared(security.WSSharedKeyRequestPreProcessors) == nil {
				ps := map[web.RequestPreProcessorName]web.RequestPreProcessor{p.Name():p}
				_ = ws.AddShared(security.WSSharedKeyRequestPreProcessors, ps)
			} else if ps, ok := ws.Shared(security.WSSharedKeyRequestPreProcessors).(map[web.RequestPreProcessorName]web.RequestPreProcessor); ok {
				if _, exists := ps[p.name]; !exists {
					ps[p.Name()] = p
				}
			}
		}
	}
	return nil
}