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

// FindStructField recursively find field that matching the given matcher, including embedded fields
func FindStructField(sType reflect.Type, matcher func(t reflect.StructField) bool) (ret reflect.StructField, found bool) {
	// dereference pointers and check type
	t := sType
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {}
	if t.Kind() != reflect.Struct {
		return ret, false
	}

	// go through fields
	for i := t.NumField() - 1; i >=0; i-- {
		f := t.Field(i)
		if f.Anonymous {
			// inspect embedded fields
			if sub, ok := FindStructField(f.Type, matcher); ok {
				sub.Index = append(f.Index, sub.Index...)
				return sub, true
			}
		} else if ok := matcher(f); ok {
			return f, true
		}
	}
	return
}
