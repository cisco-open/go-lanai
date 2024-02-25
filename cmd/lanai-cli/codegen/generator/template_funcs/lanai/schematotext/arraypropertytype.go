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

import "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/lanai/lanaiutil"

type arrayType struct {
	data                  interface{}
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func NewArrayType(data interface{}, opts ...func(option *translatorOptions)) *arrayType {
	o := &translatorOptions{}
	for _, fn := range opts {
		fn(o)
	}

	return &arrayType{
		data:                  data,
		defaultObjectName:     o.defaultObjectName,
		currentPkg:            o.currentPkg,
		restrictExternalTypes: o.restrictExternalTypes,
	}
}

func (a arrayType) ToText() (string, error) {
	schema, err := lanaiutil.ConvertToSchemaRef(a.data)
	if err != nil {
		return "", err
	}
	inner, err := NewDataTypeToTextTranslator(
		schema.Value.Items,
		WithCurrentPackage(a.currentPkg),
		WithDefaultObjectName(a.defaultObjectName),
		WithRestrictExternalTypes(a.restrictExternalTypes)).ToText()
	if err != nil {
		return "", err
	}
	return "[]" + inner, nil
}
