package utils

import "context"

// MutableContext matches *gin.Context
type MutableContext interface {
	context.Context
	Set(key string, value interface{})
}

type ListableContext interface {
	context.Context
	Values() map[interface{}]interface{}
}

type ContextValuer func(key interface{}) interface{}

// mutableContext implements MutableContext and ListableContext
type mutableContext struct {
	context.Context
	values  map[interface{}]interface{}
	valuers []ContextValuer
}

func (ctx mutableContext) Value(key interface{}) (ret interface{}) {
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

func NewMutableContext() MutableContext {
	return &mutableContext{
		Context: context.Background(),
		values: make(map[interface{}]interface{}),
	}
}

func (ctx mutableContext) Values() map[interface{}]interface{} {
	return ctx.values
}

func MakeMutableContext(parent context.Context, valuers ...ContextValuer) MutableContext {
	if mutable, ok := parent.(MutableContext); ok && len(valuers) == 0 {
		return mutable
	}

	return &mutableContext{
		Context: parent,
		values:  make(map[interface{}]interface{}),
		valuers: valuers,
	}
}
