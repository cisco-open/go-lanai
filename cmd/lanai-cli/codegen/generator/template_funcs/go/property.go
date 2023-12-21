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

// Property represents strictly what is needed to represent a struct property in Golang
type Property struct {
	Name     string
	Type     string
	Bindings string
}

func NewProperty(name string, typeOfProperty string) *Property {
	return &Property{
		Name: name,
		Type: typeOfProperty,
	}
}

func (m *Property) AddBinding(binding string) *Property {
	m.Bindings = binding
	return m
}

func (m *Property) AddType(typeOfProperty string) *Property {
	m.Type = typeOfProperty
	return m
}
