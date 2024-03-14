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

package dbtest

import (
    "embed"
    appconfig "github.com/cisco-open/go-lanai/pkg/appconfig/init"
    "github.com/cisco-open/go-lanai/pkg/data"
    "github.com/cisco-open/go-lanai/pkg/data/tx"
    "github.com/cisco-open/go-lanai/test"
    "github.com/cisco-open/go-lanai/test/apptest"
    "github.com/cisco-open/go-lanai/test/suitetest"
    "github.com/cockroachdb/copyist"
    "go.uber.org/fx"
)

//var logger = log.New("T.DB")

//go:embed defaults-dbtest.yml
var defaultConfigFS embed.FS

// EnableDBRecordMode Force enables DB recording mode.
// Normally recording mode should be enabled via `go test` argument `-record`
// IMPORTANT: when Record mode is enabled, all tests executing SQL against actual database.
// Or if Opensearch is being used, any queries to that will be executed against the real opensearch service.
// 			  So use this mode on LOCAL DEV ONLY, and have the DB copied before executing
func EnableDBRecordMode() suitetest.PackageOptions {
	return suitetest.Setup(func() error {
		setCopyistModeFlag(modeRecord)
		return nil
	})
}

// WithDBPlayback enables DB SQL playback capabilities supported by `copyist`
// This mode requires apptest.Bootstrap to work, and should not be used together with WithNoopMocks
// Each top-level test should have corresponding recorded SQL responses in `testdata` folder, or the test will fail.
// To enable record mode, use `go test ... -record` at CLI, or do it programmatically with EnableDBRecordMode
// See https://github.com/cockroachdb/copyist for more details
func WithDBPlayback(dbName string, opts ...DBOptions) test.Options {
	testOpts := withDB(modeAuto, dbName, opts)
	testOpts = append(testOpts, withData()...)
	return test.WithOptions(testOpts...)
}

// IsRecording returns true if copyist is in recording mode
func IsRecording() bool {
	return copyist.IsRecording()
}

// WithNoopMocks create a noop tx.TxManager and a noop gorm.DB
// This mode requires apptest.Bootstrap to work, and should not be used together with WithDBPlayback
// Note: in this mode, gorm.DB's DryRun and SkipDefaultTransaction are enabled
func WithNoopMocks() test.Options {
	testOpts := withData()
	testOpts = append(testOpts, apptest.WithFxOptions(
		fx.Provide(provideNoopTxManager),
		fx.Provide(provideNoopGormDialector),
		fx.Invoke(enableGormDryRun),
	))
	return test.WithOptions(testOpts...)
}

func withData() []test.Options {
	return []test.Options{
		apptest.WithModules(data.Module, tx.Module),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(pqErrorTranslatorProvider()),
		),
	}
}
