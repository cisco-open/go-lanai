// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
		// SuppressWarnings go:S108 empty block is intended
	}
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
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
		// SuppressWarnings go:S108 empty block is intended
	}
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
				ret = append(ret, sub...)
			}
		}
	}
	return
}