package bootstrap

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"

// fxPrinter implements fx.Printer
type fxPrinter struct {
	logger log.Logger
}

func (p fxPrinter) Printf(s string, v ...interface{}) {
	logger.WithContext(applicationContext).Infof(s, v...)
}

