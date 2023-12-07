package tlsconfig

import (
    "context"
    "encoding/json"
    "fmt"
    "sync"
)

type DefaultManager struct {
    sync.Mutex
    Properties       Properties
    ConfigLoaderFunc func(target interface{}, configPath string) error
    factories        map[SourceType]SourceFactory
    activeSources    map[SourceType][]Source
}

func NewDefaultManager(opts ...func(mgr *DefaultManager)) *DefaultManager {
    mgr := DefaultManager{
        factories:     make(map[SourceType]SourceFactory),
        activeSources: make(map[SourceType][]Source),
    }
    for _, fn := range opts {
        fn(&mgr)
    }
    return &mgr
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
    factory, ok := m.factories[srcCfg.Type]
    if !ok {
        return nil, fmt.Errorf("unsupported TLS source: %s", srcCfg.Type)
    }
    return factory.LoadAndInit(ctx, func(src *SourceConfig) {
        src.RawConfig = srcCfg.RawConfig
    })
}

func (m *DefaultManager) register(item interface{}) error {
    switch v := item.(type) {
    case SourceFactory:
        m.factories[v.Type()] = v
    default:
        return fmt.Errorf("unable to register unsupported item: %T", item)
    }
    return nil
}

func (m *DefaultManager) resolveSourceConfig(opt *Option) (*sourceConfig, error) {
    var src sourceConfig
    switch {
    case len(opt.Preset) != 0 && len(opt.ConfigPath) == 0 && opt.RawConfig == nil:
        preset, ok := m.Properties.Presets[opt.Preset]
		if !ok {
			return nil, fmt.Errorf(`invalid certificate options: preset [%s] is not found`, opt.Preset)
		}
		src.RawConfig = preset
    case len(opt.Preset) == 0 && len(opt.ConfigPath) != 0 && opt.RawConfig == nil:
		if e := m.ConfigLoaderFunc(&src, opt.ConfigPath); e != nil {
			return nil, fmt.Errorf(`unable to resolve certificate source configuration: %v`, e)
		}
    case len(opt.Preset) == 0 && len(opt.ConfigPath) == 0 && opt.RawConfig != nil:
        var rawJson []byte
        switch v := opt.RawConfig.(type) {
        case json.RawMessage:
            rawJson = v
        case []byte:
            rawJson = v
        case string:
            rawJson = []byte(v)
        default:
            var e error
            if rawJson, e = json.Marshal(opt.RawConfig); e != nil {
                return nil, fmt.Errorf(`invalid certificate options, unsupported RawConfig type [%T]: %v`, opt.RawConfig, e)
            }
        }
        if e := json.Unmarshal(rawJson, &src); e != nil {
            return nil, fmt.Errorf(`invalid certificate options, cannot parse "raw config" as a valid JSON block: %v`, e)
        }
        if len(opt.Type) != 0 {
            src.Type = opt.Type
        }
        return &src, nil
    default:
        return nil, fmt.Errorf(`invalid certificate options, one of "preset", "config path" or "raw config" is required. Got %v`, opt)
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
