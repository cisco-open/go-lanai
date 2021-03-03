package bootstrap

import (
	"context"
)

const (
	propertyKeyApplicationName = "application.name"
)
type LifecycleHandler func(context.Context) error

type ApplicationConfig interface {
	Value(key string) interface{}
	Bind(target interface{}, prefix string) error
}

// A Context carries addition data for application.
// delegates all other context calls to the embedded Context.
type ApplicationContext struct {
	context.Context
	config ApplicationConfig
}

func NewApplicationContext() *ApplicationContext {
	return &ApplicationContext{
		Context: context.Background(),
	}
}

func (c *ApplicationContext) Config() ApplicationConfig {
	return c.config
}

func (c *ApplicationContext) Name() string {
	name := c.Value("application.name")
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
***************************/
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