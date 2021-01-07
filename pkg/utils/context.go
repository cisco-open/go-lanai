package utils

import "context"

type MutableContext interface {
	context.Context
	SetValue(key interface{}, value interface{}) MutableContext
}

type mutableContext struct {
	context.Context
	values map[interface{}]interface{}
}

func (ctx *mutableContext) Value(key interface{}) (ret interface{}) {
	// get value from value map first, in case the key-value pair is overwritten
	ret = ctx.values[key]
	if ret == nil {
		ret = ctx.Context.Value(key)
	}
	return
}

func (ctx *mutableContext) SetValue(key interface{}, value interface{}) MutableContext {
	if key != nil && value != nil {
		ctx.values[key] = value
	}
	return ctx
}

func NewMutableContext() MutableContext {
	return &mutableContext{
		Context: context.Background(),
		values: make(map[interface{}]interface{}),
	}
}

func MakeMutableContext(parent context.Context) MutableContext {
	if mutable, ok := parent.(MutableContext); ok {
		return mutable
	}

	return &mutableContext{
		Context: parent,
		values: make(map[interface{}]interface{}),
	}
}
