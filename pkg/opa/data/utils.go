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

package opadata

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/opa"
    "gorm.io/gorm"
    "reflect"
)



/*********************
	Model Resolver
 *********************/

// policyTarget collected information about current policy target.
// The target could be a model struct instance, pointer, map of key-value pairs, etc.
type policyTarget struct {
	meta       *Metadata
	modelPtr   reflect.Value
	modelValue reflect.Value
	model      interface{}
	valueMap   map[string]interface{}
}

// toResourceValues convert to opa.ResourceValues
// might return nil without error if there is no recognized changes
func (m policyTarget) toResourceValues() (*opa.ResourceValues, error) {
	input := map[string]interface{}{}
	switch {
	case m.modelValue.IsValid():
		// create by model struct
		for k, tagged := range m.meta.Fields {
			rv := m.modelValue.FieldByIndex(tagged.StructField.Index)
			if rv.IsValid() && !rv.IsZero() {
				input[k] = rv.Interface()
			}
		}
	case m.valueMap != nil:
		// create by model map
		for k, tagged := range m.meta.Fields {
			v, _ := m.valueMap[tagged.Name]
			if v == nil {
				v, _ = m.valueMap[tagged.DBName]
			}
			if v != nil && !reflect.ValueOf(v).IsZero() {
				input[k] = v
			}
		}
	default:
		return nil, ErrUnsupportedUsage.WithMessage(`Cannot resolve values of model`)
	}

	if len(input) == 0 {
		return nil, nil
	}
	return &opa.ResourceValues{
		ExtraData: input,
	}, nil
}

// resolvePolicyTargets resolve to be created/updated/read/deleted model values.
// depending on the operation and GORM usage, values may be extracted from Dest or ReflectValue and the extracted values
// could be struct or map
func resolvePolicyTargets(stmt *gorm.Statement, meta *Metadata, op DBOperationFlag) ([]policyTarget, error) {
	resolved := make([]policyTarget, 0, 5)
	fn := func(v reflect.Value) error {
		model := policyTarget{
			meta:  meta,
			model: v.Interface(),
		}
		switch {
		case v.Type() == reflect.PointerTo(stmt.Schema.ModelType):
			model.modelPtr = v
			model.modelValue = v.Elem()
		case v.Type() == typeGenericMap:
			model.valueMap = v.Convert(typeGenericMap).Interface().(map[string]interface{})
		default:
			return fmt.Errorf("unsupported dest model [%T]", v.Interface())
		}
		resolved = append(resolved, model)
		return nil
	}

	var e error
	switch op {
	case DBOperationFlagUpdate:
		// for update, Statement.Dest should be used instead of Statement.ReflectValue.
		// See callbacks.SetupUpdateReflectValue() (update.go)
		e = walkDest(stmt, fn)
	default:
		e = walkReflectValue(stmt, fn)
	}

	if e != nil {
		return nil, fmt.Errorf("unable to extract current model model: %v", e)
	}
	return resolved, nil
}

// walkDest is similar to callbacks.callMethod. It walkthrough statement's ReflectValue
// and call given function with the pointer of the model.
func walkDest(stmt *gorm.Statement, fn func(value reflect.Value) error) (err error) {
	rv := reflect.ValueOf(stmt.Dest)
	for rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	return walkValues(rv, fn)
}

// walkReflectValue is similar to callbacks.callMethod. It walkthrough statement's ReflectValue
// and call given function with the pointer of the model.
func walkReflectValue(stmt *gorm.Statement, fn func(value reflect.Value) error) (err error) {
	return walkValues(stmt.ReflectValue, fn)
}

// walkValues recursively walk give model, support slice, array, struct and map
func walkValues(rv reflect.Value, fn func(value reflect.Value) error) error {
	//nolint:exhaustive // we only deal with map, struct and slice
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i)
			for elem.Kind() == reflect.Pointer {
				elem = elem.Elem()
			}
			if e := walkValues(elem, fn); e != nil {
				return e
			}
		}
	case reflect.Struct:
		if !rv.CanAddr() {
			return gorm.ErrInvalidValue
		}
		return fn(rv.Addr())
	case reflect.Map:
		return fn(rv)
	}
	return nil
}
