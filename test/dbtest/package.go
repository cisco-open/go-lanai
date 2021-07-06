package dbtest

import (
	appconfig "cto-github.cisco.com/NFV-BU/go-lanai/pkg/appconfig/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	datainit "cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/init"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/repo"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/tx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"embed"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

//var logger = log.New("T.DB")

const (
	_ = order.Highest + iota * 100
	orderCockroachMock
)

var dbInitialized bool

//go:embed defaults-dbtest.yml
var defaultConfigFS embed.FS

func Initialize() test.Options {
	if !dbInitialized {
		// setup without embedded/mocked DB
		return test.WithOptions(
			apptest.WithModules(tx.Module),
			apptest.WithFxOptions(
				appconfig.FxEmbeddedDefaults(defaultConfigFS),
				fx.Provide(data.BindDataProperties),
				fx.Provide(provideNoopTxManager),
				fx.Provide(newEmptyGorm),
			),
		)
	}

	// setup with embedded/mocked DB
	return test.WithOptions(
		apptest.WithModules(datainit.Module, tx.Module, repo.Module),
		apptest.WithFxOptions(
			appconfig.FxEmbeddedDefaults(defaultConfigFS),
			fx.Provide(data.BindDataProperties),
		),
	)
}

func newEmptyGorm() *gorm.DB {
	return &gorm.DB{}
}