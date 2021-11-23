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
		if ok := matcher(f); ok {
			return f, true
		}
		if f.Anonymous {
			// inspect embedded fields
			if sub, ok := FindStructField(f.Type, matcher); ok {
				sub.Index = append(f.Index, sub.Index...)
				return sub, true
			}
		}
	}
	return
}

// ListStructField recursively find all fields that matching the given matcher, including embedded fields
func ListStructField(sType reflect.Type, matcher func(t reflect.StructField) bool) (ret []reflect.StructField) {
	// dereference pointers and check type
	t := sType
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {}
	if t.Kind() != reflect.Struct {
		return
	}

	// go through fields
	for i := t.NumField() - 1; i >=0; i-- {
		f := t.Field(i)
		if ok := matcher(f); ok {
			ret = append(ret, f)
		}
		if f.Anonymous {
			// inspect embedded fields
			if sub := ListStructField(f.Type, matcher); len(sub) != 0 {
				// correct index path
				for i := range sub {
					sub[i].Index = append(f.Index, sub[i].Index...)
				}
				return sub
			}
		}
	}
	return
}