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
	typeFxIn            = reflect.TypeOf(fx.In{})
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
		shouldAnnotate, numOutput, e := validateFxProviderTarget(interfaceType, target)
		if e != nil {
			panic(e)
		}

		if shouldAnnotate {
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
		} else {
			ret[i] = fx.Annotated{
				Group:  group,
				Target: target,
			}
		}
	}
	return ret
}

// best effort to valid target provider
func validateFxProviderTarget(interfaceType reflect.Type, target interface{}) (shouldAnnotate bool, effectiveNumOut int, err error) {
	t := reflect.TypeOf(target)
	if t.Kind() != reflect.Func {
		panic(fmt.Errorf("fx annotated provider target must be a function, but got %T", target))
	}

	// 1. the return types must implements Controller except the last return value
	//   1.a if the return type is not Controller, it must be suitable for annotation (i.e. it can't use fx.In)
	// 2. the last return value can be error

	isOutputValid := true
	for i := 0; i < t.NumOut(); i++ {
		rt := t.Out(i)
		if !rt.Implements(interfaceType) {
			// if it's the last return value
			if i > 0 && i == t.NumOut()-1 {
				if !isExactType(typeError, rt) {
					isOutputValid = false
					break
				}
			} else { // every return item other than the last one must implement the expected interface
				isOutputValid = false
				break
			}
		} else {
			if !isExactType(interfaceType, rt) {
				shouldAnnotate = true
			}
			effectiveNumOut++
		}
	}

	isInputValid := true
	// check if we can actually annotate
	if shouldAnnotate {
		for i := 0; i < t.NumIn(); i++ {
			it := t.In(i)
			if it.Kind() == reflect.Struct {
				for j := 0; j < it.NumField(); j++ {
					// if the input struct embeds fx.In, then we won't be able to annotate, so it's invalid
					if isExactType(it.Field(j).Type, typeFxIn) {
						isInputValid = false
						break
					}
				}
			}
		}
	}

	if !isOutputValid {
		shouldAnnotate = false
		effectiveNumOut = 0
		err = fmt.Errorf("Web registable provider must return implementation of type %s.%s, but got %v",
			interfaceType.PkgPath(), interfaceType.Name(), describeFunc(target))
	} else if !isInputValid {
		shouldAnnotate = false
		effectiveNumOut = 0
		err = fmt.Errorf("If web registable provider does not return exact type %s.%s, it must not use Fx.In, but got %v",
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

func isExactType(expected reflect.Type, t reflect.Type) bool {
	return t.PkgPath() == expected.PkgPath() && t.Name() == expected.Name()
}
