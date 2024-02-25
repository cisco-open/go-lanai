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
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/go"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"
    "path"
    "strings"
)

type defaultType struct {
	data              interface{}
	defaultObjectName string
	currentPkg        string
}

func NewDefaultType(data interface{}, opts ...func(option *translatorOptions)) *defaultType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &defaultType{
		data:              data,
		defaultObjectName: o.defaultObjectName,
		currentPkg:        o.currentPkg,
	}
}
func (s defaultType) ToText() (string, error) {
	schema, err := lanaiutil.ConvertToSchemaRef(s.data)
	if err != nil {
		return "", err
	}
	result := ""
	if schema == nil || schema.Ref == "" {
		result = s.defaultObjectName
	} else {
		result = path.Base(schema.Ref)
		refPackage, ok := _go.StructRegistry()[strings.ToLower(result)]
		if ok && refPackage != s.currentPkg {
			result = fmt.Sprintf("%v.%v", path.Base(refPackage), result)
		}
	}
	return result, nil
}
