package validation

import (
	"context"
	"encoding"
	"fmt"
	"github.com/go-playground/validator/v10"
	"regexp"
)

func Regex(pattern string) validator.FuncCtx {
	return regex(regexp.MustCompile(pattern))
}

func RegexPOSIX(pattern string) validator.FuncCtx {
	return regex(regexp.MustCompilePOSIX(pattern))
}

func regex(compiled *regexp.Regexp ) validator.FuncCtx {
	return func(_ context.Context, fl validator.FieldLevel) bool {
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
		return compiled.MatchString(str)
	}
}

