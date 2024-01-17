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

// Package consultest
// Leveraging ittest package and HTTP VCR to record and replay consul operations
package consultest

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul"
	consulinit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/consul/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/ittest"
	"go.uber.org/fx"
	"gopkg.in/dnaeon/go-vcr.v3/recorder"
	"testing"
)

/*************************
	Top-level APIs
 *************************/

type ConsulRecorderOptions func(cfg *ConsulRecorderConfig)

type ConsulRecorderConfig struct {
	HTTPVCROptions []ittest.HTTPVCROptions
}

func WithHttpPlayback(t *testing.T, opts ...ConsulRecorderOptions) test.Options {
	cfg := ConsulRecorderConfig{}
	for _, fn := range opts {
		fn(&cfg)
	}
	testOpts := []test.Options{ittest.WithHttpPlayback(t, cfg.HTTPVCROptions...)}
	testOpts = append(testOpts,
		apptest.WithModules(consulinit.Module),
		apptest.WithFxOptions(
			fx.Provide(RecordedConsulProvider()),
		),
	)
	return test.WithOptions(testOpts...)
}

/*************************
	Top-level Options
 *************************/

// HttpRecordingMode enable "recording" mode.
// IMPORTANT:	When Record mode is enabled, all sub tests interact with real Consul service.
// 				So use this mode on LOCAL DEV ONLY
// See ittest.HttpRecordingMode()
func HttpRecordingMode() ConsulRecorderOptions {
    return func(cfg *ConsulRecorderConfig) {
        cfg.HTTPVCROptions = append(cfg.HTTPVCROptions, ittest.HttpRecordingMode())
    }
}

func MoreHTTPVCROptions(opts ...ittest.HTTPVCROptions) ConsulRecorderOptions {
    return func(cfg *ConsulRecorderConfig) {
        cfg.HTTPVCROptions = append(cfg.HTTPVCROptions, opts...)
    }
}

/*************************
	Tests Setup Helpers
 *************************/

func RecordedConsulProvider() fx.Annotated {
	return fx.Annotated{
		Group:  "consul",
		Target: ConsulWithRecorder,
	}
}

func ConsulWithRecorder(recorder *recorder.Recorder) consul.Options {
	return func(cfg *consul.ClientConfig) error {
		switch {
		case cfg.Transport != nil:
			cfg.HttpClient = recorder.GetDefaultClient()
		case cfg.HttpClient != nil:
			if cfg.HttpClient.Transport != nil {
				recorder.SetRealTransport(cfg.HttpClient.Transport)
			}
			cfg.HttpClient.Transport = recorder
		default:
			cfg.HttpClient = recorder.GetDefaultClient()
		}

		return nil
	}
}
