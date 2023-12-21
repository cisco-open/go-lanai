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

package lanai

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/openapi"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"reflect"
)

type RefChecker interface {
	CountFields() int
	ContainsRef() bool
}

func containsSingularRef(element interface{}) (bool, error) {
	fieldCount := 0
	containsRef := false
	r, err := refCheckerFactory(element)
	if err != nil {
		return false, err
	}
	for _, b := range r {
		fieldCount += b.CountFields()
		containsRef = containsRef || b.ContainsRef()
	}
	return fieldCount == 1 && containsRef, nil
}

func refCheckerFactory(element interface{}) (result []RefChecker, err error) {
	switch element.(type) {
	case *openapi3.ResponseRef:
		result = append(result, Response{openapi.Response(*element.(*openapi3.ResponseRef))})
	case *openapi3.Operation:
		// Assume this is for Requests, so give requestbodies & parameters
		op := element.(*openapi3.Operation)
		if op.RequestBody != nil {
			result = append(result, RequestBody{openapi.RequestBody(*op.RequestBody)})
		}
		result = append(result, Parameters{openapi.Parameters(op.Parameters)})
	default:
		return nil, fmt.Errorf("element not supported: %v", reflect.TypeOf(element))
	}

	return result, nil
}

func isEmpty(element interface{}) (bool, error) {
	count := 0
	r, err := refCheckerFactory(element)
	if err != nil {
		return false, err
	}
	for _, b := range r {
		count += b.CountFields()
	}

	return count == 0, nil
}
