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
	"github.com/getkin/kin-openapi/openapi3"
)

type translatorOptions struct {
	defaultObjectName     string
	currentPkg            string
	restrictExternalTypes bool
}

func WithDefaultObjectName(defaultObjectName string) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.defaultObjectName = defaultObjectName
	}
}
func WithCurrentPackage(currentPkg string) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.currentPkg = currentPkg
	}
}
func WithRestrictExternalTypes(restrictExternalTypes bool) func(options *translatorOptions) {
	return func(option *translatorOptions) {
		option.restrictExternalTypes = restrictExternalTypes
	}
}

type ToTextTranslator interface {
	ToText() (string, error)
}

func NewDataTypeToTextTranslator(element interface{}, opts ...func(option *translatorOptions)) ToTextTranslator {
	schema, _ := lanaiutil.ConvertToSchemaRef(element)
	if schema == nil {
		return NewDefaultType(element, opts...)
	}
	var translator ToTextTranslator
	switch {
	case schema.Value.Type.Is(openapi3.TypeNumber), schema.Value.Type.Is(openapi3.TypeInteger), schema.Value.Type.Is(openapi3.TypeBoolean):
		translator = NewBaseType(element, opts...)
	case schema.Value.Type.Is(openapi3.TypeString):
		translator = NewStringType(element, opts...)
	case schema.Value.Type.Is(openapi3.TypeArray):
		translator = NewArrayType(element, opts...)
	case schema.Value.Type.Is(openapi3.TypeObject):
		translator = NewObjectType(element, opts...)
	default:
		translator = NewDefaultType(element, opts...)
	}
	return translator
}
