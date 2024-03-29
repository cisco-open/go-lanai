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
	"context"
	"github.com/cisco-open/go-lanai/test"
	"github.com/go-playground/validator/v10"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

/********************
	Setup
 ********************/

var validate = validator.New()

func SetupValidate() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		validate.SetTagName("binding")
		return ctx, nil
	}
}

func RegisterCaseInsensitiveOneOf() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		e := validate.RegisterValidation("enumof", CaseInsensitiveOneOf())
		return ctx, e
	}
}

/********************
	Tests
 ********************/

func TestCaseInsensitiveOneOf(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupValidate()),
		test.Setup(RegisterCaseInsensitiveOneOf()),
		test.GomegaSubTest(SubTestCaseInsensitiveOneOfStringPositive(), "StringPositive"),
		test.GomegaSubTest(SubTestCaseInsensitiveOneOfIntPositive(), "IntPositive"),
		test.GomegaSubTest(SubTestCaseInsensitiveOneOfUIntPositive(), "UIntPositive"),
		test.GomegaSubTest(SubTestCaseInsensitiveOneOfNegative(), "Negative"),
		test.GomegaSubTest(SubTestCaseInsensitiveOneOfInvalidTag(), "InvalidTag"),
	)
}

/********************
	Sub Tests
 ********************/

func SubTestCaseInsensitiveOneOfStringPositive() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var target *ToBeValidated
		var e error
		target = &ToBeValidated{
			StringVal: "cat",
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "lower case should be valid")

		target = &ToBeValidated{
			StringVal: "DOG",
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "upper case should be valid")

		target = &ToBeValidated{
			StringVal: "Cat",
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "mixed case should be valid")
	}
}

func SubTestCaseInsensitiveOneOfIntPositive() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var target *ToBeValidated
		var e error
		target = &ToBeValidated{
			IntVal: -100,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "int should be valid")

		target = &ToBeValidated{
			IntVal: 1000,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "int should be valid")
	}
}

func SubTestCaseInsensitiveOneOfUIntPositive() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var target *ToBeValidated
		var e error
		target = &ToBeValidated{
			UIntVal: 100,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "int should be valid")

		target = &ToBeValidated{
			UIntVal: 1000,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(Succeed(), "int should be valid")
	}
}

func SubTestCaseInsensitiveOneOfNegative() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var target *ToBeValidated
		var e error
		target = &ToBeValidated{
			StringVal: "Banana",
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(BeAssignableToTypeOf(validator.ValidationErrors{}), "banana is not an animal")

		target = &ToBeValidated{
			IntVal: 2012,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(BeAssignableToTypeOf(validator.ValidationErrors{}), "2012 is not good")

		target = &ToBeValidated{
			UIntVal: 2012,
		}
		e = validate.StructCtx(ctx, target)
		g.Expect(e).To(BeAssignableToTypeOf(validator.ValidationErrors{}), "2012 is not good")
	}
}

func SubTestCaseInsensitiveOneOfInvalidTag() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		var target *ToBeValidated
		target = &ToBeValidated{
			FloatVal: 19.99,
		}
		willPanic := func() { _ = validate.StructCtx(ctx, target) }
		g.Expect(willPanic).To(Panic(), "should panic if validator is used on unsupported field")
	}
}

/********************
	Test Structs
 ********************/

type ToBeValidated struct {
	StringVal string  `binding:"omitempty,enumof=CAT dog"`
	IntVal    int     `binding:"omitempty,enumof=-100 1000"`
	UIntVal   uint    `binding:"omitempty,enumof=100 1000"`
	FloatVal  float64 `binding:"omitempty,enumof=100 1000"`
}
