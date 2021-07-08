package apptest

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig"
	appconfiginit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"embed"
	"go.uber.org/fx"
	"strings"
)

/*************************
	Test Options
 *************************/

type PropertyValuerFunc func(ctx context.Context) interface{}

// WithConfigFS provides per-test config capability.
// It register an embed.FS as application config, which could override any defaults.
// the given embed.FS should contains at least one yml file
// see appconfig.FxEmbeddedApplicationAdHoc
func WithConfigFS(fs ...embed.FS) test.Options {
	opts := make([]fx.Option, len(fs))
	for i, fs := range fs {
		opts[i] = appconfiginit.FxEmbeddedApplicationAdHoc(fs)
	}
	return WithFxOptions(opts...)
}

// WithProperties provides per-test config capability.
// It registers ad-hoc test application properties. Supported format of each Key-Value pair are:
// 	- "dotted.properties=value"
// 	- "dotted.properties: value"
// 	- "dotted.properties.without.value" implies the value is "true"
func WithProperties(kvs ...string) test.Options {
	p := newTestConfigProviderWithKV(kvs)
	return WithFxOptions(appconfiginit.FxProvideApplicationAdHoc(p))
}

// WithDynamicProperties provides per-test config capability.
// It registers ad-hoc test application properties
func WithDynamicProperties(valuers map[string]PropertyValuerFunc) test.Options {
	kvMap := make(map[string]interface{})
	for k, v := range valuers {
		kvMap[k]= v
	}
	p := NewTestConfigProvider(kvMap)
	return WithFxOptions(appconfiginit.FxProvideApplicationAdHoc(p))
}

// WithConfigFxProvider provides per-test config capability.
// It takes a fx.Option (usually fx.Provide) that returns/create appconfig.Provider
// and registers it as ad-hoc test application config provider.
// Note: Use it with caution. This is an advanced use case which typically used by other utility packages.
func WithConfigFxProvider(fxProvides ...interface{}) test.Options {
	return WithFxOptions(appconfiginit.FxProvideApplicationAdHoc(fxProvides...))
}

/*************************
	appconfig.Provider
 *************************/

// testConfigProvider implement appconfig.Provider and provide pre-defined functions
type testConfigProvider struct {
	appconfig.ProviderMeta
	kvs map[string]interface{}
}

// NewTestConfigProvider is for internal usage. Export for cross-package reference
// Use WithConfigFS, WithProperties, WithDynamicProperties instead
func NewTestConfigProvider(kvs map[string]interface{}) *testConfigProvider {

	return &testConfigProvider{
		ProviderMeta: appconfig.ProviderMeta{Precedence: 0},
		kvs: kvs,
	}
}

func newTestConfigProviderWithKV(kvs []string) *testConfigProvider {
	kvMap := make(map[string]interface{})
	for _, e := range kvs {
		// we support "a.b.c=v" or "a.b.c: v" or "a.b.c" (implies a.b.c=true)
		kv := strings.SplitN(e, "=", 2)
		if len(kv) < 2 {
			kv = strings.SplitN(e, ":", 2)
		}

		k := kv[0]
		v := "true"
		if len(kv) >= 2 {
			v = kv[1]
		}
		kvMap[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}

	return NewTestConfigProvider(kvMap)
}

func (p *testConfigProvider) Name() string {
	return "test-properties"
}

func (p *testConfigProvider) Load(ctx context.Context) (err error) {
	defer func() {
		p.Loaded = err == nil
	}()

	flatSettings := make(map[string]interface{})

	for k, v := range p.kvs {
		switch val := v.(type) {
		case string:
			flatSettings[k] = utils.ParseString(val)
		case PropertyValuerFunc:
			flatSettings[k] = val(ctx)
		}
	}

	p.Settings, err = appconfig.UnFlatten(flatSettings)
	return
}


