package dbtest

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/suitetest"
	"embed"
	"go.uber.org/fx"
)

//var logger = log.New("T.DB")

//go:embed defaults-dbtest.yml
var defaultConfigFS embed.FS


// EnableDBRecordMode Force enables DB recording mode.
// Normally recording mode should be enabled via `go test` argument `-record`
// IMPORTANT: when Record mode is enabled, all tests executing SQL against actual database.
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

// WithNoopMocks create a noop tx.TxManager and a noop gorm.DB
// This mode requires apptest.Bootstrap to work, and should not be used together with No
func WithNoopMocks() test.Options {
	testOpts := withData()
	testOpts = append(testOpts, apptest.WithFxOptions(
		fx.Provide(provideNoopTxManager),
		fx.Provide(provideNoopGormDialector),
	))
	return test.WithOptions(testOpts...)
}

func withData() []test.Options {
	return []test.Options{
		apptest.WithModules(tx.Module),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(data.BindDataProperties),
			fx.Provide(data.NewGorm),
		),
	}
}