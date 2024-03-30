package mvc

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/web"
	"net/http"
	"reflect"
)

/**********************************
	Request Decoder
***********************************/

// GinBindingRequestDecoder is a web.DecodeRequestFunc utilizing gin.Context's binding capabilities.
// The decoder instantiate the object based on Metadata.request
func GinBindingRequestDecoder(s *Metadata) web.DecodeRequestFunc {
	// No need to decode
	if s.request == nil || isHttpRequestPtr(s.request) {
		return func(c context.Context, r *http.Request) (request interface{}, err error) {
			return r, nil
		}
	}
	// decode request using GinBinding
	return web.GinBindingRequestDecoder(func() interface{} {
		return instantiateByType(s.request)
	})
}

// allocate memory space of given type.
// If the given type is a pointer, the returned value is non-nil.
// Otherwise, a zero value is returned
func instantiateByType(t reflect.Type) reflect.Value {
	switch t.Kind() {
	case reflect.Ptr:
		return reflect.New(t.Elem())
	default:
		return reflect.New(t).Elem()
	}
}

