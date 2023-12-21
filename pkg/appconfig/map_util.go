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

package appconfig

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
)

// Options the flatten options.
// By default: Delimiter = "."
type Options struct {
	Delimiter string
}

// ProcessKeyFormat traverse given map in DFS fashion and apply given processor to each KV pair
func ProcessKeyFormat(nested map[string]interface{}, processor func(string, ...func(*Options)) string) (map[string]interface{}, error) {
	result, err := processKeyFormat(nested, processor)

	return result.(map[string]interface{}), err
}

func processKeyFormat(value interface{}, processor func(string, ...func(*Options)) string) (interface{}, error) {

	switch value := value.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{})
		//if empty map, can't do anything
		if reflect.DeepEqual(value, map[string]interface{}{}) {
			return result, nil
		}
		for k, v := range value {
			newKey := processor(k)
			visitedValue, fe := processKeyFormat(v, processor)
			if fe != nil {
				return nil, fe
			}
			result[newKey] = visitedValue
		}
		return result, nil
	case []interface{}:
		//if empty slice
		var result []interface{}
		if reflect.DeepEqual(value, []interface{}{}) {
			return result, nil
		}
		for _, v := range value {
			visitedValue, fe := processKeyFormat(v, processor)
			if fe != nil {
				return nil, fe
			}
			result = append(result, visitedValue)
		}
		return result, nil
	default:
		return value, nil
	}
}

func VisitEach(nested map[string]interface{}, apply func(string, interface{}) error, configures...func(*Options)) error {
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return recursiveVisit("", nested, apply, opts)
}

//the recursive visit stops at the first error
func recursiveVisit(key string, value interface{}, apply func(string, interface{}) error, opts *Options) (err error) {
	switch value := value.(type) {
	case map[string]interface{}:
		//if empty map, can't do anything
		if reflect.DeepEqual(value, map[string]interface{}{}) {
			return
		}
		for k, v := range value {
			// create new key
			newKey := k
			if key != "" {
				newKey = key + opts.Delimiter + newKey
			}
			fe := recursiveVisit(newKey, v, apply, opts)
			if fe != nil {
				err = fe
				return
			}
		}
	case []interface{}:
		//if empty slice
		if reflect.DeepEqual(value, []interface{}{}) {
			return
		}
		for i, v := range value {
			newKey := "[" + strconv.Itoa(i) + "]"
			if key != "" {
				newKey = key + newKey
			}
			fe := recursiveVisit(newKey, v, apply, opts)
			if fe != nil {
				err = fe
				return
			}
		}
	default:
		err = apply(key, value)
	}
	return
}

type UfOptions struct {
	Delimiter   string
	AppendSlice bool
}

// UnFlatten supports un-flattening keys with index like the following
//  	my-example.url[0]=https://example.com
// The indexed entries are treated like an unsorted list. The result will be a list but the order is not
// guaranteed to reflect the index order.
// A key with multiple index (a.b[0].c[0) is not supported
func UnFlatten(flat map[string]interface{}, configures...func(*UfOptions)) (nested map[string]interface{}, err error) {
	opts := &UfOptions{
		Delimiter:   ".",
		AppendSlice: true,
	}

	for _, configure := range configures {
		configure(opts)
	}

	nested = make(map[string]interface{})

	for k, v := range flat {
		temp, e := uf(k, v, opts)
		if e != nil {
			return nil, errors.Wrap(e, "cannot un-flatten due to error in key: " + k)
		}
		err = mergo.Merge(&nested, temp, func(c *mergo.Config) {
			c.AppendSlice = opts.AppendSlice
		})
		if err != nil {
			return
		}
	}

	return
}


func uf(k string, v interface{}, opts *UfOptions) (n interface{}, err error) {
	indexOccurance := 0

	n = v

	keys := strings.Split(k, opts.Delimiter)

	for i := len(keys) - 1; i >= 0; i-- {
		currKey := keys[i]
		temp := make(map[string]interface{})

		bracketLeft := strings.Index(currKey, "[")
		bracketRight := strings.Index(currKey, "]")

		if bracketLeft > 0 && bracketRight == len(currKey) -1 {
			index, e := strconv.Atoi(currKey[bracketLeft+1 : bracketRight])
			if e != nil || index < 0 {
				return nil, errors.Wrap(e, "key:"+" has index marker [], but the index is not valid integer.")
			} else if indexOccurance > 0 {
				return nil, errors.New("key:"+" has multiple index marker []. This is not supported")
			} else {
				currKey = currKey[0:bracketLeft]
				temp[currKey] = []interface{}{n}
				indexOccurance = indexOccurance + 1
			}
		} else {
			temp[currKey] = n
		}

		n = temp
	}
	return n, nil
}

func UnFlattenKey(k string, configures...func(*Options)) []string {
	opts := &Options{
		Delimiter: ".",
	}

	for _, configure := range configures {
		configure(opts)
	}

	return strings.Split(k, opts.Delimiter)
}

const dash = rune('-')

// NormalizeKey convert camelCase key to snake-case
func NormalizeKey(key string, configures...func(*Options)) string {
	keys := UnFlattenKey(key, configures...)

	result := ""

	for i, key := range keys {
		result = result + utils.CamelToSnakeCase(key)
		if i < len(keys) - 1 {
			result = result + "."
		}
	}

	return result
}