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

package util

import (
	"errors"
	"reflect"
)

func ListContains(list []string, needle string) bool {
	for _, required := range list {
		if needle == required {
			return true
		}
	}
	return false
}

func GetInterfaceType(val interface{}) string {
	t := reflect.TypeOf(val)
	var res string
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
		res += "*"
	}
	return res + t.Name()
}

func args(values ...interface{}) []interface{} {
	return values
}

func increment(val int) int {
	return val + 1
}

func templateLog(tmpl string, args ...interface{}) string {
	logger.Debugf(tmpl, args...)
	return ""
}

func derefBoolPtr(ptr *bool) (bool, error) {
	if ptr == nil {
		return false, errors.New("pointer is nil")
	}
	return *ptr, nil
}
