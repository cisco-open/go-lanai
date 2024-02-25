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
	"github.com/cisco-open/go-lanai/pkg/log"
	"text/template"
)

var logger = log.New("Internal")
var FuncMap = template.FuncMap{
	"importsUsedByPath":   ImportsUsedByPath,
	"containsSingularRef": containsSingularRef,
	"isEmpty":             isEmpty,
	"schemaToText":        SchemaToText,
	"propertyToGoType":    PropertyToGoType,
	"shouldHavePointer":   ShouldHavePointer,
	"structTags":          structTags,
	"registerRegex":       registerRegex,
	"regex":               NewRegex,
	"versionList":         versionList,
	"mappingName":         mappingName,
	"mappingPath":         mappingPath,
	"defaultNameFromPath": defaultNameFromPath,
	"components":          NewComponents,
	"requestBody":         NewRequestBody,
	"operation":           NewOperation,
	"property":            NewProperty,
	"schema":              NewSchema,
}

func Load() {
	validatedRegexes = make(map[string]string)
}
func AddPredefinedRegexes(initialRegexes map[string]string) {
	for key, value := range initialRegexes {
		predefinedRegexes[key] = value
	}
}
