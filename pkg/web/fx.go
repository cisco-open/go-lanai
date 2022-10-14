package web

import (
	"fmt"
	"go.uber.org/fx"
	"reflect"
)

var (
	typeController      = reflect.TypeOf(func(Controller) { /* empty */ }).In(0)
	typeCustomizer      = reflect.TypeOf(func(Customizer) { /* empty */ }).In(0)
	typeErrorTranslator = reflect.TypeOf(func(ErrorTranslator) { /* empty */ }).In(0)
	typeFxOut           = reflect.TypeOf(fx.Out{})
)

func FxControllerProviders(targets ...interface{}) fx.Option {
	providers := groupedProviders(FxGroupControllers, typeController, targets)
	return fx.Provide(providers...)
}

func FxCustomizerProviders(targets ...interface{}) fx.Option {
	providers := groupedProviders(FxGroupCustomizers, typeCustomizer, targets)
	return fx.Provide(providers...)
}

func FxErrorTranslatorProviders(targets ...interface{}) fx.Option {
	providers := groupedProviders(FxGroupErrorTranslator, typeErrorTranslator, targets)
	return fx.Provide(providers...)
}

// groupedProviders construct a slice of []fx.Annotated with given "group". Basic return type checking
// is performed against expected "provideType"
func groupedProviders(group string, interfaceType reflect.Type, targets []interface{}) []interface{} {
	ret := make([]interface{}, len(targets))
	for i, target := range targets {

		ret[i] = fx.Annotate(target, fx.As(new(Controller)), fx.ResultTags(fmt.Sprintf("group:\"%s\"", FxGroupControllers)))

	}
	return ret
}
