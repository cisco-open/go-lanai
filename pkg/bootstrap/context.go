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
	"context"
	"time"
)

const (
	PropertyKeyApplicationName = "application.name"
)

type startTimeCtxKey struct{}
type stopTimeCtxKey struct{}

var (
	ctxKeyStartTime = startTimeCtxKey{}
	ctxKeyStopTime  = stopTimeCtxKey{}
)

type ApplicationConfig interface {
	Value(key string) interface{}
	Bind(target interface{}, prefix string) error
}

// ApplicationContext is a Context carries addition data for application.
// delegates all other context calls to the embedded Context.
type ApplicationContext struct {
	context.Context
	config ApplicationConfig
}

func NewApplicationContext(opts ...ContextOption) *ApplicationContext {
	ctx := context.Background()
	for _, fn := range opts {
		ctx = fn(ctx)
	}
	return &ApplicationContext{
		Context: context.WithValue(ctx, ctxKeyStartTime, time.Now().UTC()),
	}
}

func (c *ApplicationContext) Config() ApplicationConfig {
	return c.config
}

func (c *ApplicationContext) Name() string {
	name := c.Value(PropertyKeyApplicationName)
	if name == nil {
		return "lanai"
	}
	if n, ok := name.(string); ok {
		return n
	}
	return "lanai"
}

/**************************
 context.Context Interface
 **************************/

func (_ *ApplicationContext) String() string {
	return "application context"
}

func (c *ApplicationContext) Value(key interface{}) interface{} {
	if c.config == nil {
		return c.Context.Value(key)
	}

	switch key.(type) {
	case string:
		if ret := c.config.Value(key.(string)); ret != nil {
			return ret
		}
	}
	return c.Context.Value(key)
}

/**********************
* unexported methods
***********************/

func (c *ApplicationContext) withContext(parent context.Context) *ApplicationContext {
	c.Context = parent
	return c
}

func (c *ApplicationContext) withValue(k, v interface{}) *ApplicationContext {
	return c.withContext(context.WithValue(c.Context, k, v))
}
