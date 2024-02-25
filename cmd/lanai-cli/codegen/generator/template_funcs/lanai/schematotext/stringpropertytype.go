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

package schematotext

import (
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
	"github.com/google/uuid"
	"reflect"
	"time"
)

type stringType struct {
	data                 interface{}
	baseProcessor        *baseType
	externalTypesAllowed bool
}

func NewStringType(data interface{}, opts ...func(option *translatorOptions)) *stringType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &stringType{
		data:                 data,
		baseProcessor:        NewBaseType(data, opts...),
		externalTypesAllowed: !o.restrictExternalTypes,
	}
}
func (s stringType) ToText() (string, error) {
	result, err := s.baseProcessor.ToText()
	if err != nil {
		return "", err
	}
	if s.externalTypesAllowed {
		//Perform modifications on the base type depending on the format
		if lanaiutil.MatchesFormat(s.data, "uuid") {
			result = reflect.TypeOf(uuid.UUID{}).String()
		} else if lanaiutil.MatchesFormat(s.data, "date-time") {
			result = reflect.TypeOf(time.Time{}).String()
		}
	}
	return result, nil
}
