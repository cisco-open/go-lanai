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

package gomegautils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

const TestJson = `
{
    "firstName": "John",
    "lastName": "doe",
    "age": 26,
    "address": {
        "streetAddress": "naist street",
        "city": "Nara",
        "postalCode": "630-0192"
    },
    "phoneNumbers": [
        {
            "type": "iPhone",
            "number": "0123-4567-8888"
        },
        {
            "type": "home",
            "number": "0123-4567-8910"
        }
    ]
}
`

func TestJsonPathMatchers(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestHaveJsonPath(), "TestHaveJsonPath"),
		test.GomegaSubTest(SubTestHaveJsonPathWithValue(), "TestHaveJsonPathWithValue"),
		test.GomegaSubTest(SubTestFailureMessages(), "TestFailureMessages"),
	)
}

func SubTestHaveJsonPath() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(TestJson).To(HaveJsonPath("$.firstName"))
		g.Expect(TestJson).To(HaveJsonPath("$..type"))
		g.Expect(TestJson).NotTo(HaveJsonPath("$.type"))
	}
}

func SubTestHaveJsonPathWithValue() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(TestJson).To(HaveJsonPathWithValue("$.firstName", ContainElements("John")))
		g.Expect(TestJson).To(HaveJsonPathWithValue("$.lastName", "doe"))
		g.Expect(TestJson).To(HaveJsonPathWithValue("$..type", HaveLen(2)))
		g.Expect(TestJson).NotTo(HaveJsonPathWithValue("$.type", ContainElement("Android")))
	}
}

func SubTestFailureMessages() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		matcher := HaveJsonPathWithValue("$.lastName", "doe")
		msg := matcher.FailureMessage([]byte(TestJson))
		g.Expect(msg).To(Not(BeEmpty()))
		msg = matcher.NegatedFailureMessage(TestJson)
		g.Expect(msg).To(Not(BeEmpty()))
	}
}
