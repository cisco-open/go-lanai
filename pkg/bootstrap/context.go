package bootstrap

import (
	"context"
)

type LifecycleHandler func(context.Context) error

// A Context carries addition data for application.
// delegates all other context calls to the embedded Context.
type Context struct {
	context.Context
	Environment map[string]interface{}
}

func NewContext() *Context {
	return &Context{
		Context: context.Background(),
		Environment: make(map[string]interface{}),
	}
}

/**************************
 context.Context Interface
***************************/
func (c *Context) UpdateParent(parent context.Context) *Context {
	c.Context = parent
	return c
}

func (c *Context) Value(key interface{}) interface{} {
	return c.Environment[key.(string)]
}

func (_ *Context) String() string {
	return "bootstrap context"
}

func (c *Context) PutValue(key string, value interface{}) {
	c.Environment[key] = value
}
