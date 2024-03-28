package web

import (
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
)

/**********************************
	Request Decoder
***********************************/

// GinBindingRequestDecoder is a DecodeRequestFunc utilizing gin.Context's binding capabilities.
// The decoder uses the provided function to instantiate the object.
// If the instantiateFunc returns a non-pointer value, the decoder uses reflect to find its pointer
func GinBindingRequestDecoder(instantiateFunc func() interface{}) DecodeRequestFunc {
	// decode request using gin.Context's bind functions
	return func(c context.Context, r *http.Request) (request interface{}, err error) {
		ginCtx := GinContext(c)
		if ginCtx == nil {
			return nil, NewHttpError(http.StatusInternalServerError, errors.New("context issue"))
		}

		toBind, toRet := resolveBindable(instantiateFunc())

		// We always try to bind H, Uri and Query. other bindings are determined by Content-Type (in ShouldBind)
		err = bind(toBind,
			ginCtx.ShouldBindHeader,
			ginCtx.ShouldBindUri,
			ginCtx.ShouldBindQuery)

		if err != nil {
			return nil, translateBindingError(err)
		}

		err = ginCtx.ShouldBind(toBind)

		if err != nil && !(errors.Is(err, io.EOF) && r.ContentLength <= 0) {
			return nil, translateBindingError(err)
		}
		return toRet.Interface(), validateBinding(c, toBind)
	}
}

type bindingFunc func(interface{}) error

func bind(obj interface{}, bindings ...bindingFunc) (err error) {
	for _, b := range bindings {
		if err = b(obj); err != nil {
			return
		}
	}
	return
}

func translateBindingError(err error) error {
	return NewBindingError(err)
}

func validateBinding(ctx context.Context, obj interface{}) error {
	if bindingValidator == nil {
		return nil
	}

	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	switch v.Kind() {
	case reflect.Struct:
		return bindingValidator.StructCtx(ctx, obj)
	default:
		return nil
	}
}

// resolveBindable using reflection to resolve bindable pointer of actual value.
func resolveBindable(i interface{}) (bindablePtr interface{}, actual reflect.Value) {
	switch v := i.(type) {
	case reflect.Value:
		actual = v
	default:
		actual = reflect.ValueOf(i)
	}
	switch actual.Kind() {
	case reflect.Ptr:
		bindablePtr = actual.Interface()
	default:
		if actual.CanAddr() {
			bindablePtr = actual.Addr().Interface()
		}
	}
	return
}


