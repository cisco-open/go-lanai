package template

import (
	"context"
	"net/http"
	"reflect"
)

type RequestContext map[string]interface{}

// MakeRequestContext collect http.Request's exported fields and additional context values
func MakeRequestContext(ctx context.Context, r *http.Request, contextKeys...string) RequestContext {
	rc := RequestContext{}
	rval := reflect.ValueOf(r).Elem()
	rtype := rval.Type()
	for i := rtype.NumField() - 1; i >= 0; i-- {
		f := rtype.Field(i)
		if f.PkgPath == "" && f.Type.Kind() != reflect.Func {
			// TODO we should filter the values
			// we only put exported fields
			v := rval.Field(i).Interface()
			rc[f.Name] = v
		}
	}

	for _,key := range contextKeys {
		v := ctx.Value(key)
		if v != nil {
			rc[key] = v
		}
	}
	return rc
}
