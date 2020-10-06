package bootstrap

import "time"

type BootstrapContext struct {
	Environment map[string]interface{}
}

func (*BootstrapContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (*BootstrapContext) Done() <-chan struct{} {
	return nil
}

func (*BootstrapContext) Err() error {
	return nil
}

func (ctx *BootstrapContext) Value(key interface{}) interface{} {
	return ctx.Environment[key.(string)]
}

func (e *BootstrapContext) String() string {
	return "Bootstrap context"
}

func (ctx *BootstrapContext) PutValue(key string, value interface{}) {
	ctx.Environment[key] = value
}
