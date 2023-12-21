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

package generator

/*********************
	Project
 *********************/

type Project struct {
	Name        string
	Module      string
	Description string
	Port        int
	ContextPath string
}

/*********************
	Components
 *********************/

type Components struct {
	Contract Contract
	Security Security
}

/*********************
	API Contract
 *********************/

type Contract struct {
	Path   string
	Naming ContractNaming
}

type ContractNaming struct {
	RegExps map[string]string
}

/*********************
	Web Security
 *********************/

type Security struct {
	Authentication Authentication
	Access         Access
}

type AuthenticationMethod string

const (
	AuthNone   AuthenticationMethod = `none`
	AuthOAuth2 AuthenticationMethod = `oauth2`
	// TODO more authentication methods like basic, form, etc...
)

type Authentication struct {
	Method AuthenticationMethod
}

type AccessPreset string

const (
	AccessPresetFreestyle AccessPreset = `freestyle`
	AccessPresetOPA       AccessPreset = `opa`
)

type Access struct {
	Preset AccessPreset
}

/******************
	Regen
 ******************/

// RegenMode file operation mode when re-generating.
type RegenMode string

const (
	RegenModeIgnore    RegenMode = "ignore"
	RegenModeReference RegenMode = "reference"
	RegenModeOverwrite RegenMode = "overwrite"
)

// RegenRule file operation rules during re-generation
type RegenRule struct {
	// Pattern wildcard pattern of output file path
	Pattern string
	// Mode regeneration mode on matched output files in case of changes. (ignore, overwrite, reference, etc.)
	Mode RegenMode
}

type RegenRules []RegenRule
