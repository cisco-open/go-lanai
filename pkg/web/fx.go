package web

import (
	"fmt"
	"go.uber.org/fx"
	"reflect"
	"runtime"
)

var (
	typeController      = reflect.TypeOf(func(Controller) { /* empty */ }).In(0)
	typeCustomizer      = reflect.TypeOf(func(Customizer) { /* empty */ }).In(0)
	typeErrorTranslator = reflect.TypeOf(func(ErrorTranslator) { /* empty */ }).In(0)
	typeError           = reflect.TypeOf((*error)(nil)).Elem()
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
		numOutput, e := validateFxProviderTarget(interfaceType, target)
		if e != nil {
			panic(e)
		}

		types := make([]interface{}, numOutput)
		tags := make([]string, numOutput)
		for i := 0; i < numOutput; i++ {
			//The fx.As(interfaces ...interface{}) expects pointer to interface, i.e. fx.As(new(io.Writer)).
			// So if we want to annotate something as Controller, we need to initialize a *Controller variable to use in fx.As.
			// Here interfaceType is Controller,
			// so reflect.New will give us a Value variable representing a pointer to zero value of Controller, in other words, a *Controller.
			// Then the Interface() call goes from Value to interface{} so that we can use it in fx.As(interfaces ...interface{})
			types[i] = reflect.New(interfaceType).Interface()
			tags[i] = fmt.Sprintf("group:\"%s\"", group)
		}
		annotation := fx.As(types...)

		ret[i] = fx.Annotate(target, annotation, fx.ResultTags(tags...))
	}
	return ret
}

// best effort to valid target provider
func validateFxProviderTarget(interfaceType reflect.Type, target interface{}) (effectiveNumOut int, err error) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("fx annotated provider target must be a function, but got %T", target))
	}

	// 1. the return types must be homogenous except the last return value
	// 2. the last return value can be error

	isValid := true
	for i := 0; i < t.NumOut(); i++ {
		rt := t.Out(i)
		if !isSupportedType(interfaceType, rt) {
			// if it's the last return value
			if i > 0 && i == t.NumOut()-1 {
				if !isSupportedType(typeError, rt) {
					isValid = false
					break
				}
			} else { // every return item other than the last one must implement the expected interface
				isValid = false
				break
			}
		} else {
			effectiveNumOut++
		}
	}

	if !isValid {
		effectiveNumOut = 0
		err = fmt.Errorf("Web registable provider must return type %s.%s, but got %v",
			interfaceType.PkgPath(), interfaceType.Name(), describeFunc(target))
	} else {
		err = nil
	}
	return
}

func describeFunc(f interface{}) string {
	pc := reflect.ValueOf(f).Pointer()
	pFunc := runtime.FuncForPC(pc)
	if pFunc == nil {
		return "unknown function"
	}
	return pFunc.Name()
}

func isSupportedType(expected reflect.Type, t reflect.Type) bool {
	return t.Implements(expected)
}
