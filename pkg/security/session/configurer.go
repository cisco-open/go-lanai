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

package session

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/bootstrap"
    "github.com/cisco-open/go-lanai/pkg/security"
    "github.com/cisco-open/go-lanai/pkg/security/session/common"
    "github.com/cisco-open/go-lanai/pkg/web/middleware"
)

var (
	FeatureId = security.FeatureId("Session", security.FeatureOrderSession)
)

// Feature holds session configuration
type Feature struct {
	sessionName    string
	settingService SettingService
}

func (f *Feature) Identifier() security.FeatureIdentifier {
	return FeatureId
}

func (f *Feature) SettingService(settingService SettingService) *Feature {
	f.settingService = settingService
	return f
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
	store        Store
	sessionProps security.SessionProperties
}

func newSessionConfigurer(sessionProps security.SessionProperties, sessionStore Store) *Configurer {
	return &Configurer{
		store:        sessionStore,
		sessionProps: sessionProps,
	}
}

func (sc *Configurer) Apply(feature security.Feature, ws security.WebSecurity) error {
	f := feature.(*Feature)

	if len(f.sessionName) == 0 {
		f.sessionName = common.DefaultName
	}

	// the ws shared store is to share this store with other feature configurer can have access to store.
	if ws.Shared(security.WSSharedKeySessionStore) == nil {
		_ = ws.AddShared(security.WSSharedKeySessionStore, sc.store)
	}

	// configure middleware
	manager := NewManager(f.sessionName, sc.store)

	sessionHandler := middleware.NewBuilder("sessionMiddleware").
		Order(security.MWOrderSessionHandling).
		Use(manager.SessionHandlerFunc())

	authPersist := middleware.NewBuilder("authPersistMiddleware").
		Order(security.MWOrderAuthPersistence).
		Use(manager.AuthenticationPersistenceHandlerFunc())

	//test := middleware.NewBuilder("post-sessionMiddleware").
	//	Order(security.MWOrderAuthPersistence + 10).
	//	Use(SessionDebugHandlerFunc())

	ws.Add(sessionHandler, authPersist)

	// configure auth success/error handler
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(&ChangeSessionHandler{})
	if bootstrap.DebugEnabled() {
		ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
			Add(&DebugAuthSuccessHandler{})
		ws.Shared(security.WSSharedKeyCompositeAuthErrorHandler).(*security.CompositeAuthenticationErrorHandler).
			Add(&DebugAuthErrorHandler{})
	}

	var settingService SettingService
	if f.settingService == nil {
		settingService = NewDefaultSettingService(sc.sessionProps)
	} else {
		settingService = f.settingService
	}

	concurrentSessionHandler := &ConcurrentSessionHandler{
		sessionStore:          sc.store,
		sessionSettingService: settingService,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(concurrentSessionHandler)

	deleteSessionHandler := &DeleteSessionOnLogoutHandler{
		sessionStore: sc.store,
	}
	ws.Shared(security.WSSharedKeyCompositeAuthSuccessHandler).(*security.CompositeAuthenticationSuccessHandler).
		Add(deleteSessionHandler)
	return nil
}
