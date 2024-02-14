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

package utils

import (
	"context"
)

// MutableContext wraps context.Context with an internal KV pairs storage.
// KV pairs stored in this context can be changed in later time.
type MutableContext interface {
	context.Context
	Set(key string, value interface{})
}
// ListableContext is supplementary interface of MutableContext, listing all values stored in the context
type ListableContext interface {
	context.Context
	Values() map[interface{}]interface{}
}

// ExtendedMutableContext additional interface, can set KV without wrapping new context
type ExtendedMutableContext interface {
	MutableContext
	SetKV(key interface{}, value interface{})
}

type ContextValuer func(key interface{}) interface{}

// ckMutableContext is the key for itself
type mutableContextKey struct{}
var ckMutableContext = mutableContextKey{}

// mutableContext implements GinContext, ListableContext and MutableContext
type mutableContext struct {
	context.Context
	values  map[interface{}]interface{}
	valuers []ContextValuer
}

func (ctx *mutableContext) Value(key interface{}) (ret interface{}) {
	switch key {
	case ckMutableContext:
		return ctx
	}

	// get value from value map first, in case the key-value pair is overwritten
	ret, ok := ctx.values[key]
	if !ok || ret == nil {
		ret = ctx.Context.Value(key)
	}

	if ret == nil && ctx.valuers != nil {
		// use valuers to get
		for _, valuer := range ctx.valuers {
			if ret = valuer(key); ret != nil {
				return
			}
		}
	}

	return
}

func (ctx *mutableContext) Set(key string, value interface{}) {
	if key != "" && value != nil {
		ctx.values[key] = value
	} else if key != "" {
		delete(ctx.values, key)
	}
}

func (ctx *mutableContext) SetKV(key interface{}, value interface{}) {
	if key != nil && value != nil {
		ctx.values[key] = value
	} else if key != nil {
		delete(ctx.values, key)
	}
}

func (ctx *mutableContext) Values() map[interface{}]interface{} {
	return ctx.values
}

func NewMutableContext() MutableContext {
	return &mutableContext{
		Context: context.Background(),
		values:  make(map[interface{}]interface{}),
	}
}

// MakeMutableContext return the context itself if it's already a MutableContext and no additional ContextValuer are specified.
// Otherwise, wrap the given context as MutableContext.
// Note: If given context hierarchy contains MutableContext as parent context, their mutable store (map) is shared
func MakeMutableContext(parent context.Context, valuers ...ContextValuer) MutableContext {
	if mutable, ok := parent.(*mutableContext); ok && len(valuers) == 0 {
		return mutable
	}
	return &mutableContext{
		Context: parent,
		values:  make(map[interface{}]interface{}),
		valuers: valuers,
	}
}

// FindMutableContext return MutableContext from given context.Context's inheritance hierarchy.
// Important: this function may returns parent context of the given one. T
//			  Therefore, the returned context is only for mutating KV pairs and SHOULD NOT be passed along.
func FindMutableContext(ctx context.Context) MutableContext {
	mc, _ := ctx.Value(ckMutableContext).(MutableContext)
	return mc
}
