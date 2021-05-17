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
	Load() error
	// GetSettings returns loaded settings. might be nil if not IsLoaded returns true
	// The returned map should be un-flattened. i.e. flat.key=value should be stored as {"flat":{"key":"value"}}
	GetSettings() map[string]interface{}
	// IsLoaded should return true if Load is invoked at least once
	IsLoaded() bool
	// Reset delete loaded settings and reset IsLoaded flag
	Reset()
}

type ProviderReorderer interface {
	// Reorder set receivers
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
