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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"fmt"
	opalogging "github.com/open-policy-agent/opa/logging"
	"github.com/open-policy-agent/opa/plugins"
	"github.com/open-policy-agent/opa/sdk"
	"io"
)

var embeddedOPA struct {
	opa              *sdk.OPA
	ready            EmbeddedOPAReadyCH
	inputCustomizers []InputCustomizer
}

type EmbeddedOPAReadyCH <-chan struct{}

func EmbeddedOPA() *sdk.OPA {
	return embeddedOPA.opa
}

type EmbeddedOPAOptions func(opts *EmbeddedOPAOption)
type EmbeddedOPAOption struct {
	// SDKOptions raw sdk.Options
	SDKOptions sdk.Options
	// Config struct overrides SDKOptions.Config
	Config *Config
	// InputCustomizers installed as global input customizers for any OPA queries
	InputCustomizers []InputCustomizer
	// Properties for extra configuration that not included in Config
	Properties *Properties
}

func WithConfig(cfg *Config) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.Config = cfg
	}
}

func WithRawConfig(jsonReader io.Reader) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.SDKOptions.Config = jsonReader
	}
}

func WithLogger(logger opalogging.Logger) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.SDKOptions.Logger = logger
	}
}

func WithLogLevel(level log.LoggingLevel) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.SDKOptions.Logger = NewOPALogger(logger, level)
	}
}

func WithInputCustomizers(customizers ...InputCustomizer) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.InputCustomizers = customizers
	}
}

func WithProperties(props Properties) EmbeddedOPAOptions {
	return func(opts *EmbeddedOPAOption) {
		opts.Properties = &props
	}
}

// NewEmbeddedOPA create a new sdk.OPA instance and make it available via EmbeddedOPA function.
// Caller is responsible to call (*sdk.OPA).Stop to release resources
func NewEmbeddedOPA(ctx context.Context, opts ...EmbeddedOPAOptions) (*sdk.OPA, EmbeddedOPAReadyCH, error) {
	readyCh := make(chan struct{}, 1)
	opt := EmbeddedOPAOption{
		SDKOptions: sdk.Options{
			ID:    `Embedded-OPA`,
			Ready: readyCh,
			Plugins: map[string]plugins.Factory{
				pluginNameDecisionLogger: decisionLogPluginFactory{},
			},
		},
	}
	for _, fn := range opts {
		fn(&opt)
	}
	if e := validateOptions(ctx, &opt); e != nil {
		return nil, nil, e
	}

	opa, e := sdk.New(ctx, opt.SDKOptions)
	if e != nil {
		close(readyCh)
		return nil, nil, fmt.Errorf("error when create embedded OPA: %v", e)
	}
	// set global variable
	embeddedOPA.opa = opa
	embeddedOPA.ready = readyCh
	embeddedOPA.inputCustomizers = opt.InputCustomizers
	return opa, readyCh, nil
}

func validateOptions(ctx context.Context, opt *EmbeddedOPAOption) error {
	// check logger
	if opt.SDKOptions.Logger == nil {
		opaLog := NewOPALogger(logger.WithContext(ctx), log.LevelInfo)
		WithLogger(opaLog)(opt)
	} else if v, ok := opt.SDKOptions.Logger.(*opaLogger); ok {
		WithLogger(v.WithContext(ctx))(opt)
	}

	// check config
	switch {
	case opt.Config == nil && opt.SDKOptions.Config == nil:
		return fmt.Errorf(`"Config" is missing`)
	case opt.Config != nil:
		reader, e := opt.Config.JSONReader(ctx)
		if e != nil {
			return e
		}
		WithRawConfig(reader)(opt)
	}
	return nil
}
