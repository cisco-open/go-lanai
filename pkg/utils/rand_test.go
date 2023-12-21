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

package utils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestRandomInteger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestRandomInt64N(), "RandomInt64N"),
		test.GomegaSubTest(SubTestRandomIntN(), "RandomIntN"),
	)
}

func TestRandomString(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestRandomString(), "Alphanumeric"),
		test.GomegaSubTest(SubTestRandomStringWithCharset(), "CustomCharset"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestRandomInt64N() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const maxInt64 = int64(^uint64(0) >> 1)
		const max = 100
		r := int64(-1)
		r = RandomInt64N(maxInt64)
		g.Expect(r).To(BeNumerically(">=", 0), "random int64 should be greater than 0")

		r = RandomInt64N(max)
		g.Expect(r).To(BeNumerically(">=", 0), "random int64 should be greater than 0")
		g.Expect(r).To(BeNumerically("<", max), "random int64 should be less than %d", max)
	}
}

func SubTestRandomIntN() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const maxInt = int(^uint(0) >> 1)
		const max = 100
		r := -1
		r = RandomIntN(maxInt)
		g.Expect(r).To(BeNumerically(">=", 0), "random int should be greater than 0")

		r = RandomIntN(max)
		g.Expect(r).To(BeNumerically(">=", 0), "random int should be greater than 0")
		g.Expect(r).To(BeNumerically("<", max), "random int should be less than %d", max)
	}
}

func SubTestRandomString() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const regexAlphanumeric = `^[a-zA-Z0-9]{0,%d}$`
		const length = 100
		r := RandomString(length)
		g.Expect(r).To(HaveLen(length), "random string should have correct length")
		g.Expect(r).To(MatchRegexp(regexAlphanumeric, length), "random string should be alphanumeric")

		r = RandomString(-1)
		g.Expect(r).To(BeEmpty(), "random string with non-positive length should be empty")
	}
}

func SubTestRandomStringWithCharset() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const charset = "13579asdfghjkl"
		const regex = `^[13579asdfghjkl]{0,%d}$`
		const length = 100
		r := RandomStringWithCharset(length, charset)
		g.Expect(r).To(HaveLen(length), "random string should have correct length")
		g.Expect(r).To(MatchRegexp(regex, length), "random string should be alphanumeric")

		r = RandomStringWithCharset(-1, charset)
		g.Expect(r).To(BeEmpty(), "random string with non-positive length should be empty")
	}
}
