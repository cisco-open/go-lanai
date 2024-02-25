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
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/log"
    "github.com/open-policy-agent/opa/plugins"
    opalogs "github.com/open-policy-agent/opa/plugins/logs"
    "github.com/open-policy-agent/opa/rego"
)

var evtLogger = log.New("OPA.Event")

const (
	pluginNameDecisionLogger = `lanai_logger`
	kLogDecisionLog          = `opa`
	kLogDecisionReason = `reason`
	kLogPartialResult  = `result`
	kLogPartialReason  = `reason`
)

/*******************
	Log Context
 *******************/

type kLogCtx struct{}

var kLogCtxLevel = kLogCtx{}

type logContext struct {
	context.Context
	level log.LoggingLevel
}

func (c logContext) Value(key any) any {
	switch key {
	case kLogCtxLevel:
		return c.level
	}
	return c.Context.Value(key)
}

func logContextWithLevel(parent context.Context, level log.LoggingLevel) context.Context {
	return &logContext{
		Context: parent,
		level:   level,
	}
}

/*******************
	Leveled Log
 *******************/

// eventLogger get a logger with context and level properly configured
func eventLogger(ctx context.Context, defaultLevel log.LoggingLevel) log.Logger {
	return evtLogger.WithContext(ctx).WithLevel(resolveLogLevel(ctx, defaultLevel))
}

func resolveLogLevel(ctx context.Context, defaultLevel log.LoggingLevel) log.LoggingLevel {
	override, ok := ctx.Value(kLogCtxLevel).(log.LoggingLevel)
	if !ok {
		return defaultLevel
	}
	return override
}

/*******************
	Decision Log
 *******************/

type decisionLogPluginFactory struct{}

func (f decisionLogPluginFactory) Validate(_ *plugins.Manager, rawConfig []byte) (interface{}, error) {
	var props LoggingProperties
	if e := json.Unmarshal(rawConfig, &props); e != nil {
		return nil, e
	}
	return props, nil
}

func (f decisionLogPluginFactory) New(manager *plugins.Manager, cfg interface{}) plugins.Plugin {
	manager.UpdatePluginStatus(pluginNameDecisionLogger, &plugins.Status{
		State:   plugins.StateOK,
		Message: fmt.Sprintf("Plugin is ready [%s]", pluginNameDecisionLogger),
	})
	return &decisionLogger{
		level: cfg.(LoggingProperties).DecisionLogsLevel,
	}
}

// decisionLogger OPA SDK decision logger plugin. Implementing "github.com/open-policy-agent/opa/plugins/logs".Logger
type decisionLogger struct {
	level log.LoggingLevel
}

func (l *decisionLogger) Start(_ context.Context) error {
	return nil
}

func (l *decisionLogger) Stop(_ context.Context) {
	// does nothing
}

func (l *decisionLogger) Reconfigure(_ context.Context, cfg interface{}) {
	l.level = cfg.(LoggingProperties).DecisionLogsLevel
}

func (l *decisionLogger) Log(ctx context.Context, v1 opalogs.EventV1) error {
	eventLogger(ctx, l.level).
		WithKV(kLogDecisionLog, decisionEvent{event: &v1}).
		Printf("Decision Log")
	return nil
}

/*******************
	Events
 *******************/

type decisionEvent struct {
	event *opalogs.EventV1
}

func (de decisionEvent) String() string {
	v, e := json.Marshal(de.event)
	if e != nil {
		return fmt.Sprintf("%v", de.event)
	}
	return string(v)
}

func (de decisionEvent) MarshalJSON() ([]byte, error) {
	return json.Marshal(de.event)
}

type resultEvent struct {
	ID     string      `json:"decision_id"`
	Result interface{} `json:"result"`
	Deny   bool        `json:"deny"`
}

func (re resultEvent) String() string {
	return fmt.Sprintf("[%s]: %v", re.ID, re.Result)
}

type partialQueriesLog rego.PartialQueries

func (pq partialQueriesLog) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%v"`, rego.PartialQueries(pq))), nil
}

type partialResultEvent struct {
	ID  string             `json:"decision_id"`
	Err error              `json:"error,omitempty"`
	AST *partialQueriesLog `json:"queries,omitempty"`
}

func (pre partialResultEvent) String() string {
	if pre.Err != nil {
		return pre.Err.Error()
	}
	return fmt.Sprintf("[%s]: %v", pre.ID, pre.AST)
}
