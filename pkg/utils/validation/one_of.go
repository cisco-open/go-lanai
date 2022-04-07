package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// CaseInsensitiveOneOf validator function that similar to validator.isOneOf but case-insensitive
func CaseInsensitiveOneOf() validator.Func {
	return func(fl validator.FieldLevel) bool {
		vals := parseOneOfParam2(fl.Param())

		field := fl.Field()

		var v string
		switch field.Kind() {
		case reflect.String:
			v = field.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			v = strconv.FormatInt(field.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			v = strconv.FormatUint(field.Uint(), 10)
		default:
			panic(fmt.Sprintf("Bad field type %T", field.Interface()))
		}
		for i := 0; i < len(vals); i++ {
			if strings.EqualFold(vals[i], v) {
				return true
			}
		}
		return false
	}
}

var splitParamsRegex = regexp.MustCompile(`'[^']*'|\S+`)
var oneofValsCache = map[string][]string{}
var oneofValsCacheRWLock = sync.RWMutex{}

func parseOneOfParam2(s string) []string {
	oneofValsCacheRWLock.RLock()
	vals, ok := oneofValsCache[s]
	oneofValsCacheRWLock.RUnlock()
	if !ok {
		oneofValsCacheRWLock.Lock()
		vals = splitParamsRegex.FindAllString(s, -1)
		for i := 0; i < len(vals); i++ {
			vals[i] = strings.Replace(vals[i], "'", "", -1)
		}
		oneofValsCache[s] = vals
		oneofValsCacheRWLock.Unlock()
	}
	return vals
}



