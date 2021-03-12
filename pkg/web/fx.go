package web

import (
	"fmt"
	"go.uber.org/fx"
	"reflect"
	"runtime"
)

var (
	typeController      = reflect.TypeOf(func(Controller) {}).In(0)
	typeCustomizer      = reflect.TypeOf(func(Customizer) {}).In(0)
	typeErrorTranslator = reflect.TypeOf(func(ErrorTranslator) {}).In(0)
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
		if e := validateFxProviderTarget(interfaceType, target); e != nil {
			panic(e)
		}

		ret[i] = fx.Annotated{
			Group:  group,
			Target: target,
		}
	}
	return ret
}

// best effort to valid target provider
func validateFxProviderTarget(interfaceType reflect.Type, target interface{}) error {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("fx annotated provider target must be a function, but got %T", target))
	}

	for i := 0; i < t.NumOut(); i++ {
		rt := t.Out(i)
		if isExactType(interfaceType, rt) {
			return nil
		}
	}
	return fmt.Errorf("Web registable provider must return type %s.%s, but got %v",
		interfaceType.PkgPath(), interfaceType.Name(), describeFunc(target))
}

func describeFunc(f interface{}) string {
	pc := reflect.ValueOf(f).Pointer()
	pFunc := runtime.FuncForPC(pc)
	if pFunc == nil {
		return "unknown function"
	}
	return pFunc.Name()
}

func isExactType(expected reflect.Type, t reflect.Type) bool {
	return t.PkgPath() == expected.PkgPath() && t.Name() == expected.Name()
}