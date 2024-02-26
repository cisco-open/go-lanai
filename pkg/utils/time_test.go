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
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/test"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "testing"
    "time"
)

func TestTimeUtils(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestTimeISO8601(), "TestTimeISO8601"),
		test.GomegaSubTest(SubTestTimeCustomLayout(), "TestTimeCustomLayout"),
		test.GomegaSubTest(SubTestDuration(), "TestDuration"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestTimeISO8601() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		texts := map[string]bool{
			time.Now().Format(time.RFC3339): true,
			"2024-01-11T19:21:11Z":          true,
			"2024-01-11T12:21:11-05:00":     true,
			"2024-01-11T12:21:11.234Z":      true,
			"2024-01-11T12:21:11.234-05:00": true,
			"2024-01-11":                    false,
			"19:21:11Z":                     false,
			"T192111Z":                      false,
		}

		for text, success := range texts {
			parsed := ParseTimeISO8601(text)
			if success {
				g.Expect(parsed).ToNot(BeZero(), "parsing %s should be successful", text)
			} else {
				g.Expect(parsed).To(BeZero(), "parsing %s should fail with zero return", text)
			}
		}
	}
}

func SubTestTimeCustomLayout() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		texts := map[string]bool{
			time.Now().Format(time.RFC822): true,
			"11 Jan 24 20:40 GMT":          true,
			"11 Jan 24 15:40 EST":          true,
			"11 Jan 24 20:40 UTC":          true,
			"11 Jan 24 20:40":              false,
		}
		for text, success := range texts {
			parsed := ParseTime(time.RFC822, text)
			if success {
				g.Expect(parsed).ToNot(BeZero(), "parsing %s should be successful", text)
			} else {
				g.Expect(parsed).To(BeZero(), "parsing %s should fail with zero return", text)
			}
		}
	}
}

func SubTestDuration() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		// Valid time units are "ns", "us" (or "µs"), "ms", "s", "m", "h".
		texts := map[string]bool{
			"9h10m24s123ms321us998ns": true,
			"-9h10m123ms998ns":        true,
			"25h61m99s":               true,
			"5d9h22m":                 false,
		}

		for text, success := range texts {
			d := ParseDuration(text)
			if !success {
				g.Expect(d).To(BeZero(), "parsing %s should fail with zero return", text)
				var another Duration
				e := json.Unmarshal([]byte(fmt.Sprintf(`"%s"`, text)), &another)
				g.Expect(e).To(HaveOccurred(), "unmarshaling from %s should fail", text)
				continue
			}
			g.Expect(d).ToNot(BeZero(), "parsing %s should be successful", text)
			data, e := json.Marshal(Duration(d))
			g.Expect(e).To(Succeed(), "marshaling %v should not fail", d)
			var another Duration
			e = json.Unmarshal(data, &another)
			g.Expect(e).To(Succeed(), "unmarshaling %s should not fail", string(data))
			g.Expect(another).To(BeEquivalentTo(d), "unmarshaled value should be correct")
		}
	}
}
