package utils

import "context"

// MutableContext matches *gin.Context
type MutableContext interface {
	context.Context
	Set(key string, value interface{})
}

type mutableContext struct {
	context.Context
	values map[interface{}]interface{}
}

func (ctx *mutableContext) Value(key interface{}) (ret interface{}) {
	// get value from value map first, in case the key-value pair is overwritten
	ret, ok := ctx.values[key]
	if !ok || ret == nil {
		ret = ctx.Context.Value(key)
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
