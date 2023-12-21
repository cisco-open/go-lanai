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

package bootstrap

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"go.uber.org/fx/fxevent"
	"strings"
)

// fxPrinter implements fx.Printer (deprecated) and fxevent.Logger
type fxPrinter struct {
	logger log.Logger
	appCtx *ApplicationContext
}

func provideFxLogger(app *App) fxevent.Logger {
	return newFxPrinter(logger, app)
}

func newFxPrinter(logger log.Logger, app *App) *fxPrinter {
	return &fxPrinter{
		logger: logger,
		appCtx: app.ctx,
	}
}

func (l *fxPrinter) logf(msg string, args ...interface{}) {
	logger.WithContext(l.appCtx).Infof(msg, args...)
}

func (l *fxPrinter) Printf(s string, v ...interface{}) {
	logger.WithContext(l.appCtx).Infof(s, v...)
}

func (l *fxPrinter) LogEvent(event fxevent.Event) {
	switch e := event.(type) {
	case *fxevent.OnStartExecuting:
		//logger.WithContext(l.appCtx).Debugf("HOOK OnStart\t\t%s executing (caller: %s)", e.FunctionName, e.CallerName)
	case *fxevent.OnStartExecuted:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("HOOK OnStart\t\t%s called by %s failed in %s: %v", e.FunctionName, e.CallerName, e.Runtime, e.Err)
		} //else {
		//logger.WithContext(l.appCtx).Debugf("HOOK OnStart\t\t%s called by %s ran successfully in %s", e.FunctionName, e.CallerName, e.Runtime)
		//}
	case *fxevent.OnStopExecuting:
		logger.WithContext(l.appCtx).Debugf("HOOK OnStop\t\t%s executing (caller: %s)", e.FunctionName, e.CallerName)
	case *fxevent.OnStopExecuted:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("HOOK OnStop\t\t%s called by %s failed in %s: %v", e.FunctionName, e.CallerName, e.Runtime, e.Err)
		} //else {
		//logger.WithContext(l.appCtx).Debugf("HOOK OnStop\t\t%s called by %s ran successfully in %s", e.FunctionName, e.CallerName, e.Runtime)
		//}
	case *fxevent.Supplied:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\tFailed to supply %v: %v", e.TypeName, e.Err)
		} else {
			logger.WithContext(l.appCtx).Infof("SUPPLY\t%v", e.TypeName)
		}
	case *fxevent.Provided:
		for _, rtype := range e.OutputTypeNames {
			logger.WithContext(l.appCtx).Infof("PROVIDE\t%v <= %v", rtype, e.ConstructorName)
		}
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("Error after options were applied: %v", e.Err)
		}
	case *fxevent.Invoking:
		logger.WithContext(l.appCtx).Debugf("INVOKE\t\t%s", e.FunctionName)
	case *fxevent.Invoked:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\t\tfx.Invoke(%v) called from:\n%+vFailed: %v", e.FunctionName, e.Trace, e.Err)
		}
	case *fxevent.Stopping:
		logger.WithContext(l.appCtx).Infof("%v", strings.ToUpper(e.Signal.String()))
	case *fxevent.Stopped:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\t\tFailed to stop cleanly: %v", e.Err)
		}
	case *fxevent.RollingBack:
		logger.WithContext(l.appCtx).Warnf("ERROR\t\tStart failed, rolling back: %v", e.StartErr)
	case *fxevent.RolledBack:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\t\tCouldn't roll back cleanly: %v", e.Err)
		}
	case *fxevent.Started:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\t\tFailed to start: %v", e.Err)
		} else {
			logger.WithContext(l.appCtx).Infof("RUNNING")
		}
	case *fxevent.LoggerInitialized:
		if e.Err != nil {
			logger.WithContext(l.appCtx).Warnf("ERROR\t\tFailed to initialize custom logger: %+v", e.Err)
		} else {
			logger.WithContext(l.appCtx).Infof("LOGGER\tInitialized custom logger from %v", e.ConstructorName)
		}
	}
}
