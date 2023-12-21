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

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
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
		Properties:    *NewProperties(),
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
	src, e := factory.LoadAndInit(ctx, func(src *SourceConfig) {
		src.RawConfig = srcCfg.RawConfig
	})
	if e != nil {
		return nil, e
	}

	sources, _ := m.activeSources[srcCfg.Type]
	sources = append(sources, src)
	m.activeSources[srcCfg.Type] = sources
	return src, nil
}

func (m *DefaultManager) Close() error {
	for _, sources := range m.activeSources {
		for _, src := range sources {
			if closer, ok := src.(io.Closer); ok {
				_ = closer.Close()
			}
		}
	}
	return nil
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
		if e := json.Unmarshal(preset, &src); e != nil {
			return nil, fmt.Errorf(`unable to resolve certificate source preset [%s]: %v`, opt.Preset, e)
		}
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
