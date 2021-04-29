package validation

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/security"
	"encoding"
	"fmt"
	"github.com/go-playground/validator/v10"
)

func TenantAccess() validator.FuncCtx {
	return func(ctx context.Context, fl validator.FieldLevel) bool {
		i := fl.Field().Interface()
		var str string
		switch v := i.(type) {
		case string:
			str = v
		case *string:
			if v != nil {
				str = *v
			}
		case fmt.Stringer:
			str = v.String()
		case encoding.TextMarshaler:
			bytes, _ := v.MarshalText()
			str = string(bytes)
		default:
			// we don't validate non string, just fail it
			return false
		}
		return security.HasAccessToTenant(ctx, str)
	}
}

