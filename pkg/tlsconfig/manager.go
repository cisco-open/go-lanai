package tlsconfig

import (
    "context"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
    "encoding/json"
    "fmt"
    "sync"
)

type DefaultManager struct {
    sync.Mutex
	AppConfig bootstrap.ApplicationConfig
	Factories map[SourceType]SourceFactory
    Providers map[string]Provider
}

func NewDefaultManager(appCfg bootstrap.ApplicationConfig) *DefaultManager {
	return &DefaultManager{
        AppConfig: appCfg,
		Factories: make(map[SourceType]SourceFactory),
        Providers:  make(map[string]Provider),
	}
}

func (m *DefaultManager) Register(items ...interface{}) error {
	for _, item := range items {
		if e := m.register(item); e != nil {
			return e
		}
	}
	return nil
}

func (m *DefaultManager) MustRegister(items ...interface{}) {
	if e := m.Register(items...); e != nil {
		panic(e)
	}
}

func (m *DefaultManager) Source(ctx context.Context, opts ...Options) (Source, error) {
	opt := Option{}
	for _, fn := range opts {
		fn(&opt)
	}
	srcCfg, e := m.resolveSourceConfig(&opt)
	if e != nil {
		return nil, e
	}

	m.Lock()
	defer m.Unlock()
	factory, ok := m.Factories[srcCfg.Type]
	if !ok {
		return nil, fmt.Errorf("unsupported TLS source: %s", srcCfg.Type)
	}
	return factory.LoadAndInit(ctx, func(src *SourceConfig) {
		src.RawConfig = srcCfg.RawConfig
	})
}

// Provider
// Deprecated
func (m *DefaultManager) Provider(ctx context.Context, opts ...Options) (Provider, error) {
	srcFactory, e := m.Source(ctx, opts...)
	return srcFactory.(Provider), e
}

func (m *DefaultManager) register(item interface{}) error {
	switch v := item.(type) {
	case SourceFactory:
		m.Factories[v.Type()] = v
	default:
		return fmt.Errorf("unable to register unsupported item: %T", item)
	}
	return nil
}

func (m *DefaultManager) resolveSourceConfig(opt *Option) (*sourceConfig, error) {
	var src sourceConfig
    switch {
    case len(opt.Preset) != 0 && len(opt.ConfigPath) == 0 && len(opt.Type) == 0:
        opt.ConfigPath = fmt.Sprintf("%s.presets.%s", PropertiesPrefix, opt.Preset)
    case len(opt.Preset) == 0 && len(opt.ConfigPath) != 0 && len(opt.Type) == 0:
        // do nothing
    case len(opt.Preset) == 0 && len(opt.ConfigPath) == 0 && len(opt.Type) != 0:
		src.Type = opt.Type
		switch v := opt.RawConfig.(type) {
		case json.RawMessage:
			src.RawConfig = v
		case []byte:
			src.RawConfig = v
		case string:
			src.RawConfig = []byte(v)
		default:
			raw, e := json.Marshal(opt.RawConfig)
			if e != nil {
				return nil, fmt.Errorf(`invalid certificate options, unsupported RawConfig type [%T]: %v`, opt.RawConfig, e)
			}
			src.RawConfig = raw
		}
		return &src, nil
    default:
        return nil, fmt.Errorf(`invalid certificate options, "preset", "config path" and "raw config" are exclusive. Got %v`, opt)
    }

    if e := m.AppConfig.Bind(&src, opt.ConfigPath); e != nil {
        return nil, fmt.Errorf(`unable to resolve certificate source configuration: %v`, e)
    }
    return &src, nil
}

/*************************
	Helpers
 *************************/

type sourceConfig struct {
	Type      SourceType      `json:"type"`
	RawConfig json.RawMessage `json:"-"`
}

func (c *sourceConfig) UnmarshalJSON(data []byte) error {
	c.RawConfig = data
	type cfg sourceConfig
	return json.Unmarshal(data, (*cfg)(c))
}