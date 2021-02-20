package reflectutils

import (
	"reflect"
	"unicode"
)

func IsExportedField(f reflect.StructField) bool {
	if len(f.Name) == 0 {
		return false
	}
	r := rune(f.Name[0])
	return unicode.IsUpper(r)
}
