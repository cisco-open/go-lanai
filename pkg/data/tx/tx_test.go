package tx

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"cto-github.cisco.com/NFV-BU/go-lanai/test/apptest"
	"database/sql"
	"github.com/onsi/gomega"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"testing"
)

type noopTxManager struct {}

func provideNoopTxManager() TxManager {
	return noopTxManager{}
}

func (m noopTxManager) Transaction(_ context.Context, _ TxFunc, _ ...*sql.TxOptions) error {
	return nil
}

func (m noopTxManager) WithDB(_ *gorm.DB) GormTxManager {
	return m
}

/*************************
	Tests
 *************************/

func TestOverridingTxManager(t *testing.T) {
	test.RunTest(context.Background(), t,
		apptest.Bootstrap(),
		apptest.WithModules(Module),
		apptest.WithFxOptions(
			fx.Provide(provideNoopTxManager),
		),
		test.GomegaSubTest(SubTestNoopTxManager(), "VerifyNoopTxManager"),
	)
}

// TODO more tests

/*************************
	Sub Tests
 *************************/

func SubTestNoopTxManager() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(txManager).To(gomega.BeAssignableToTypeOf(noopTxManager{}))
		e := Transaction(context.Background(), func(ctx context.Context) error {
			return nil
		})
		g.Expect(e).To(gomega.Succeed(), "TxManager shouldn't return error")
	}
}