package bootstrap

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"

// fxPrinter implements fx.Printer
type fxPrinter struct {
	logger log.Logger
	appCtx *ApplicationContext
}

func newFxPrinter(logger log.Logger, app *App) *fxPrinter {
	return &fxPrinter{
		logger: logger,
		appCtx: app.ctx,
	}
}

func (p fxPrinter) Printf(s string, v ...interface{}) {
	logger.WithContext(p.appCtx).Infof(s, v...)
}

