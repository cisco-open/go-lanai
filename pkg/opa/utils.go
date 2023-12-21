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

package opa

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// marshalMergedJSON merge extra into v, v have to be struct or map.
// "processors" are invoked after merge and before marshalling
func marshalMergedJSON(obj interface{}, extra map[string]interface{}, processors ...func(m map[string]interface{})) ([]byte, error) {
	data, e := json.Marshal(obj)
	if len(extra) == 0 && len(processors) == 0 || e != nil {
		return data, e
	}
	// merge extra
	var m map[string]interface{}
	if e := json.Unmarshal(data, &m); e != nil {
		return nil, fmt.Errorf("unable to merge JSON: %v", e)
	}
	for k, v := range extra {
		m[k] = v
	}
	for _, fn := range processors {
		fn(m)
	}
	return json.Marshal(m)
}

// minimizeMap recursively remove any zero valued entries
func minimizeMap(m map[string]interface{}) {
	minimize(reflect.ValueOf(m))
}

func minimize(rv reflect.Value) (minimized reflect.Value, isZero bool) {
	if rv.Kind() == reflect.Interface {
		rv = rv.Elem()
	}
	isZero = !rv.IsValid() || rv.IsZero()
	//nolint:exhaustive // we only deal with map and slice
	switch rv.Kind() {
	case reflect.Map:
		rv = minimizeMapValue(rv)
		isZero = isZero || rv.Len() == 0
	case reflect.Slice:
		rv = minimizeSliceValue(rv)
		isZero = isZero || rv.Len() == 0
	}
	return rv, isZero
}

func minimizeMapValue(mapV reflect.Value) reflect.Value {
	for _, k := range mapV.MapKeys() {
		v := mapV.MapIndex(k)
		v, zero := minimize(v)
		if zero {
			mapV.SetMapIndex(k, reflect.Value{})
		} else {
			mapV.SetMapIndex(k, v)
		}
	}
	return mapV
}

func minimizeSliceValue(sliceV reflect.Value) reflect.Value {
	newV := reflect.MakeSlice(sliceV.Type().Elem(), 0, sliceV.Len())
	for i:=0; i < sliceV.Len(); i ++ {
		v := sliceV.Index(i)
		if v, zero := minimize(v); !zero {
			newV = reflect.Append(newV, v)
		}
	}
	return newV
}

