package certs

func WithSourceProperties(props *SourceProperties) Options {
    return func(opt *Option) {
        if len(props.Preset) != 0 {
            opt.Preset = props.Preset
        } else {
            opt.RawConfig = props.Raw
        }
    }
}

func WithPreset(presetName string) Options {
    return func(opt *Option) {
        opt.Preset = presetName
    }
}

func WithConfigPath(configPath string) Options {
    return func(opt *Option) {
        opt.ConfigPath = configPath
    }
}

func WithRawConfig(rawCfg interface{}) Options {
    return func(opt *Option) {
        opt.RawConfig = rawCfg
    }
}

func WithType(srcType SourceType, cfg interface{}) Options {
    return func(opt *Option) {
        opt.Type = srcType
        opt.RawConfig = cfg
    }
}
