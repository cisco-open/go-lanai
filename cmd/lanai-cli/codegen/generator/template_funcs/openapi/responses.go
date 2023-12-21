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

package openapi

import (
	"github.com/getkin/kin-openapi/openapi3"
	"sort"
)

type Responses openapi3.Responses

func (r Responses) RefsUsed() (result []string) {
	for _, response := range r.Sorted() {
		// check if a struct will be created from this response
		if response.CountFields() == 0 || (response.CountFields() == 1 && response.ContainsRef()) {
			break
		}
		result = append(result, response.RefsUsed()...)
	}
	return result
}
func (r Responses) Sorted() (result []Response) {
	keys := make([]string, 0, len(r))
	for k := range r {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result = append(result, Response(*r[k]))
	}

	return result
}
