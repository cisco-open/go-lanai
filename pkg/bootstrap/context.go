package bootstrap

import (
	"context"
	"cto-github.cisco.com/livdu/jupiter/pkg/appconfig"
	"fmt"
)

type LifecycleHandler func(context.Context) error

// A Context carries addition data for application.
// delegates all other context calls to the embedded Context.
type ApplicationContext struct {
	context.Context
	config *appconfig.Config
}

func NewContext() *ApplicationContext {
	return &ApplicationContext{
		Context: context.Background(),
	}
}

/**************************
 context.Context Interface
***************************/
func (c *ApplicationContext) UpdateConfig(config *appconfig.Config) {
	c.config = config
}

func (c *ApplicationContext) UpdateParent(parent context.Context) *ApplicationContext {
	c.Context = parent
	return c
}

func (_ *ApplicationContext) String() string {
	return "application context"
}

func (c *ApplicationContext) Value(key interface{}) interface{} {
	//TODO: This method is meant to be accessed only after application context is loaded completely
	// PANIC if this method is called before fully ready

	value, error := c.config.Value(key.(string))

	if error == nil {
		return value
	}

	return nil
}

func (c *ApplicationContext) dumpConfigurations() {
	c.config.Each(func(key string, value interface{}) {
		fmt.Println(key + ": " + fmt.Sprintf("%v", value))
	})
}