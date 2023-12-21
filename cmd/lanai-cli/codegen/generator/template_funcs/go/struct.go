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

package _go

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs/util"
)

type Struct struct {
	Properties      []Property
	EmbeddedStructs []Struct
	Package         string
	Name            string
}

func NewStruct(name string, pkg string) *Struct {
	if name == "" {
		return nil
	}
	return &Struct{
		Name:    name,
		Package: pkg,
	}
}

func (m *Struct) AddProperties(property Property) *Struct {
	if m == nil {
		return nil
	}
	m.Properties = append(m.Properties, property)
	return m
}

func (m *Struct) AddEmbeddedStruct(ref string) *Struct {
	if m == nil {
		return nil
	}

	name := util.ToTitle(util.BasePath(ref))
	m.EmbeddedStructs = append(m.EmbeddedStructs, Struct{
		Package: StructLocation(name),
		Name:    name,
	})

	return m
}
