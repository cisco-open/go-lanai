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

package appconfig

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

type Provider interface {
	order.Ordered

	// Name is unique name of given provider, it also used as primary key in any mapping
	Name() string
	// Load load settings and should be idempotent. e.g. calling it multiple times should not affect loaded settings
	Load(ctx context.Context) error
	// GetSettings returns loaded settings. might be nil if not IsLoaded returns true
	// The returned map should be un-flattened. i.e. flat.key=value should be stored as {"flat":{"key":"value"}}
	GetSettings() map[string]interface{}
	// IsLoaded should return true if Load is invoked at least once
	IsLoaded() bool
	// Reset delete loaded settings and reset IsLoaded flag
	Reset()
}

type ProviderReorderer interface {
	// Reorder set order
	Reorder(int)
}

// ProviderGroup determines Providers based on given bootstrap.ApplicationConfig
type ProviderGroup interface {
	order.Ordered

	// Providers returns providers based on given config.
	// This method should be idempotent. e.g. calling it multiple times with same config always returns identical slice
	Providers(ctx context.Context, config bootstrap.ApplicationConfig) []Provider

	// Reset should mark all providers unloaded
	Reset()
}

/********************
	Common Impl.
 ********************/

// ProviderMeta implements ProviderReorderer and partial ProviderMeta
type ProviderMeta struct {
	Loaded     bool                   //invalid if not loaded or during load
	Settings   map[string]interface{} //storage for the settings loaded by the auth
	Precedence int                    //the precedence for which the settings will take effect.
}

func (m ProviderMeta) GetSettings() map[string]interface{} {
	return m.Settings
}

func (m ProviderMeta) Order() int {
	return m.Precedence
}

func (m ProviderMeta) IsLoaded() bool {
	return m.Loaded
}

func (m *ProviderMeta) Reset() {
	m.Loaded = false
	m.Settings = nil
}

func (m *ProviderMeta) Reorder(order int) {
	m.Precedence = order
}
