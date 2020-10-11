package bootstrap

import (
	"context"
	"time"
)

type LifecycleHandler func(context.Context) error

type Context struct {
	Environment map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		Environment: make(map[string]interface{}),
	}
}

/**************************
 context.Context Interface
***************************/
func (*Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*Context) Done() <-chan struct{} {
	return nil
}

func (*Context) Err() error {
	return nil
}

func (ctx *Context) Value(key interface{}) interface{} {
	return ctx.Environment[key.(string)]
}

func (e *Context) String() string {
	return "Bootstrap context"
}

func (ctx *Context) PutValue(key string, value interface{}) {
	ctx.Environment[key] = value
}
