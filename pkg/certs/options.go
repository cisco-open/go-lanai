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
