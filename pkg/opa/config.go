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

package opa

import (
	"bytes"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"encoding/json"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/open-policy-agent/opa/download"
	opakeys "github.com/open-policy-agent/opa/keys"
	"github.com/open-policy-agent/opa/plugins/bundle"
	opadiscovery "github.com/open-policy-agent/opa/plugins/discovery"
	opalogs "github.com/open-policy-agent/opa/plugins/logs"
	oparest "github.com/open-policy-agent/opa/plugins/rest"
	opastatus "github.com/open-policy-agent/opa/plugins/status"
	opacache "github.com/open-policy-agent/opa/topdown/cache"
	"io"
	"net/url"
	"time"
)

const defaultServerName = `opa-bundle-service`

type ConfigCustomizer interface {
	Customize(ctx context.Context, cfg *Config)
}

// Config is a subset OPA Config with typed field
// see OPA's Config.Config and Config.ParseConfig
type Config struct {
	Services                     map[string]*oparest.Config `json:"services,omitempty"`
	Labels                       map[string]string          `json:"labels,omitempty"`
	Discovery                    *opadiscovery.Config       `json:"discovery,omitempty"`
	Bundles                      map[string]*bundle.Source  `json:"bundles,omitempty"`
	DecisionLogs                 *opalogs.Config            `json:"decision_logs,omitempty"`
	Status                       *opastatus.Config          `json:"status,omitempty"`
	Plugins                      map[string]interface{}     `json:"plugins,omitempty"`
	Keys                         map[string]*opakeys.Config `json:"keys,omitempty"`
	DefaultDecision              *string                    `json:"default_decision,omitempty"`
	DefaultAuthorizationDecision *string                    `json:"default_authorization_decision,omitempty"`
	Caching                      *opacache.Config           `json:"caching,omitempty"`
	NDBuiltinCache               bool                       `json:"nd_builtin_cache,omitempty"`
	PersistenceDirectory         *string                    `json:"persistence_directory,omitempty"`
	DistributedTracing           *distributedTracingConfig  `json:"distributed_tracing,omitempty"`
	Storage                      *storageConfig             `json:"storage,omitempty"`
	ExtraConfig                  map[string]interface{}     `json:"-"`
}

func (c Config) MarshalJSON() ([]byte, error) {
	type config Config
	return marshalMergedJSON(config(c), c.ExtraConfig, minimizeMap)
}

func (c Config) MarshalText() ([]byte, error) {
	return json.Marshal(c)
}

func (c Config) JSONReader(ctx context.Context) (io.Reader, error) {
	var buf bytes.Buffer
	if e := json.NewEncoder(&buf).Encode(&c); e != nil {
		return nil, e
	}
	logger.WithContext(ctx).Debugf("OPA Config: %s", buf.Bytes())
	return &buf, nil
}

// see OPA's internal distributedtracing.distributedTracingConfig (internal/distributedtracing/distributedtracing.go)
type distributedTracingConfig struct {
	Type                  string `json:"type,omitempty"`
	Address               string `json:"address,omitempty"`
	ServiceName           string `json:"service_name,omitempty"`
	SampleRatePercentage  *int   `json:"sample_percentage,omitempty"`
	EncryptionScheme      string `json:"encryption,omitempty"`
	EncryptionSkipVerify  *bool  `json:"allow_insecure_tls,omitempty"`
	TLSCertFile           string `json:"tls_cert_file,omitempty"`
	TLSCertPrivateKeyFile string `json:"tls_private_key_file,omitempty"`
	TLSCACertFile         string `json:"tls_ca_cert_file,omitempty"`
}

type storageConfig struct {
	Disk *diskConfig `json:"disk,omitempty"`
}

// see OPA's disk.cfg (disk/Config.go)
type diskConfig struct {
	Dir        string   `json:"directory"`
	AutoCreate bool     `json:"auto_create"`
	Partitions []string `json:"partitions"`
	Badger     string   `json:"badger"`
}

// LoadConfig create config and combine values from defaults and properties
func LoadConfig(ctx context.Context, props Properties, customizers ...ConfigCustomizer) (*Config, error) {
	cfg := Config{
		Plugins:     map[string]interface{}{},
		ExtraConfig: map[string]interface{}{},
	}
	if e := applyProperties(&props, &cfg); e != nil {
		return nil, e
	}
	for _, customizer := range customizers {
		customizer.Customize(ctx, &cfg)
	}
	return &cfg, nil
}

func applyProperties(props *Properties, cfg *Config) error {
	// service
	serverName := props.Server.Name
	if len(serverName) == 0 {
		serverName = defaultServerName
	}
	if _, e := url.Parse(props.Server.URL); e != nil {
		return fmt.Errorf(`invalid OPA server URL: %v`, e)
	}
	cfg.Services = map[string]*oparest.Config{
		serverName: {
			Name: serverName,
			URL:  props.Server.URL,
			//AllowInsecureTLS:             true,
		},
	}

	// decision logs
	if props.Logging.DecisionLogsLevel != log.LevelOff {
		cfg.Plugins[pluginNameDecisionLogger] = props.Logging
		cfg.DecisionLogs = &opalogs.Config{
			Plugin:      utils.ToPtr(pluginNameDecisionLogger),
		}
	}

	// bundles
	cfg.Bundles = map[string]*bundle.Source{}
	for k := range props.Bundles {
		v := props.Bundles[k]
		polling := props.Server.PollingProperties
		if e := mergo.Merge(&polling, &v.PollingProperties, mergo.WithOverride); e != nil {
			return fmt.Errorf("unable to merge polling properties of bundle [%s]: %v", k, e)
		}
		cfg.Bundles[k] = &bundle.Source{
			Config: download.Config{
				Trigger: nil,
				Polling: download.PollingConfig{
					MinDelaySeconds:           asSeconds(polling.PollingMinDelay),
					MaxDelaySeconds:           asSeconds(polling.PollingMaxDelay),
					LongPollingTimeoutSeconds: asSeconds(polling.LongPollingTimeout),
				},
			},
			Service:  serverName,
			Resource: v.Path,
		}
	}
	return nil
}

func asSeconds(duration *utils.Duration) *int64 {
	if duration == nil {
		return nil
	}
	secs := int64(time.Duration(*duration) / time.Second)
	return &secs
}
