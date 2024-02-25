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
// To change/list KV pairs on any context.Context that inherit from MutableContext,
// use FindMutableContext to obtain a MutableContextAccessor.
// See FindMutableContext and MutableContextAccessor for more details
type MutableContext interface {
	context.Context
	Set(key, value any)
}
// ListableContext is supplementary interface of MutableContext, listing all values stored in the context
type ListableContext interface {
	context.Context
	Values() map[interface{}]interface{}
}

// ContextValuer is an additional source of context.Context.Value(any)) used by MutableContext to search values with key.
// When MutableContext cannot find given key in its internal store, it will go through all ContextValuers
// before pass along the key-value searching to its parent context.
// See NewMutableContext and MakeMutableContext
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
	if ok && ret != nil {
		return
	}

	// use valuers to get
	for _, valuer := range ctx.valuers {
		if ret = valuer(key); ret != nil {
			return
		}
	}

	// pass along to parent
	return ctx.Context.Value(key)
}

func (ctx *mutableContext) Set(key any, value any) {
	if key != nil && value != nil {
		ctx.values[key] = value
	} else if key != nil {
		delete(ctx.values, key)
	}
}

// Values recursively gather all KVs stored in MutableContext and its parent contexts.
// In case of overridden keys, the value of outermost context is used.
func (ctx *mutableContext) Values() (values map[interface{}]interface{}) {
	hierarchy := make([]*mutableContext, 0, 5)
	for mc := ctx; mc != nil; mc, _ = mc.Context.Value(ckMutableContext).(*mutableContext) {
		hierarchy = append(hierarchy, mc)
	}

	// go over the inheritance hierarchy from root to current, in case the value is overridden
	values = make(map[interface{}]interface{})
	for i := len(hierarchy) - 1; i >= 0; i-- {
		for k, v := range hierarchy[i].values {
			values[k] = v
		}
	}
	return values
}

// NewMutableContext Wrap given context.Context with a mutable store and optionally additional KV sources defined as ContextValuer
//nolint:contextcheck // false positive - Non-inherited new context, use function like `context.WithXXX` instead
func NewMutableContext(parent context.Context, valuers ...ContextValuer) MutableContext {
	if parent == nil {
		parent = context.Background()
	}
	return &mutableContext{
		Context: parent,
		values:  make(map[interface{}]interface{}),
		valuers: valuers,
	}
}

// MakeMutableContext return the context itself if it's already a MutableContext and no additional ContextValuer are specified.
// Otherwise, wrap the given context as MutableContext.
// Note: If the given context itself is not a MutableContext but its hierarchy contains MutableContext as parent context,
// 		 a new MutableContext is created and its mutable store (map) is not shared with the one from the hierarchy.
func MakeMutableContext(parent context.Context, valuers ...ContextValuer) MutableContext {
	if mutable, ok := parent.(*mutableContext); ok && len(valuers) == 0 {
		return mutable
	}
	return NewMutableContext(parent, valuers...)
}

type MutableContextAccessor interface {
	Set(key, value any)
	Values() (values map[any]any)
}

// FindMutableContext search for MutableContext from given context.Context's inheritance hierarchy,
// and return a MutableContextAccessor for key-values manipulation.
// If MutableContext is not found, nil is returned.
//
// Important: This function may returns parent context of the given one. Therefore, changing values may affect parent context.
func FindMutableContext(ctx context.Context) MutableContextAccessor {
	if mc, ok := ctx.Value(ckMutableContext).(*mutableContext); ok {
		return mc
	}
	return nil
}
