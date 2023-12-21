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

package web

import (
	"fmt"
	"github.com/gin-gonic/gin/binding"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

var (
	bindingValidator = newValidator(binding.Validator)
)

// Validator returns the global validator for binding.
// Callers can register custom validators
func Validator() *Validate {
	return bindingValidator
}


func newValidator(ginValidator binding.StructValidator) *Validate {
	validate := ginValidator.Engine().(*validator.Validate)
	return &Validate{
		Validate: validate,
	}
}

// Validate is a thin wrapper around validator/v10, which prevent modifying TagName
type Validate struct {
	*validator.Validate
}

// WithTagName create a shallow copy of internal validator.Validate with different tag name
func (v *Validate) WithTagName(name string) *Validate {
	cp := Validate{
		Validate: v.Validate,
	}
	cp.Validate.SetTagName(name)
	return &cp
}

func (v *Validate) SetTagName(name string) {
	panic(fmt.Errorf("illegal attempt to modify tag of validator. Please use WithTagName(string)"))
}

// SetTranslations registers default translations using given regFn
func (v *Validate) SetTranslations(trans ut.Translator, regFn func(*validator.Validate, ut.Translator) error) error {
	return regFn(v.Validate, trans)
}

