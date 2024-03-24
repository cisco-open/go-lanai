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

package validation

import (
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

const (
	DefaultLocale = "en"
)

var (
	universalTranslator = newUniversalTranslator()
)

// DefaultTranslator returns the default ut.Translator of the package
func DefaultTranslator() ut.Translator {
	trans, _ := universalTranslator.GetTranslator(DefaultLocale)
	return trans
}

// UniversalTranslator returns the globally configured ut.UniversalTranslatorTranslator
// callers can register more locales
func UniversalTranslator() *ut.UniversalTranslator {
	return universalTranslator
}

// SimpleTranslationRegFunc returns a translation registration function for simple validation translate template
// the returned function could be used to register custom translation override
func SimpleTranslationRegFunc(tag, template string) func(*validator.Validate, ut.Translator) error {
	return func(validate *validator.Validate, trans ut.Translator) error {
		return validate.RegisterTranslation(tag, trans, func(ut ut.Translator) error {
			return ut.Add(tag, template, true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T(tag, fe.Field(), fe.Param())
			return t
		})
	}
}

func newUniversalTranslator() *ut.UniversalTranslator {
	english := en.New()
	return ut.New(english)
}
