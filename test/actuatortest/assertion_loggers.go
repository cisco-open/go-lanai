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

package actuatortest

import (
	. "cto-github.cisco.com/NFV-BU/go-lanai/test/utils/gomega"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"net/http"
	"testing"
)

type ExpectedLoggersOptions func(l *ExpectedLoggers)
type ExpectedLoggers struct {
	Single           bool
	EffectiveLevels  map[string]interface{}
	ConfiguredLevels map[string]interface{}
}

// ExpectLoggersSingleEntry expects logger response is a single entry.
// Maximum of two "expected levels" are supported:
// 	- 	1st level is expected "effective level".
//		If nil or missing, it's expected to be any value.
// 	- 	2nd level is expected "configured level".
//		If nil, it's expected to be any value.
//		If missing, it's not checked at all
// Note: "effectiveLevel" is expected always available in any "loggers" response.
func ExpectLoggersSingleEntry(expectedLevels...interface{}) ExpectedLoggersOptions {
	return func(l *ExpectedLoggers) {
		l.Single = true
		l.EffectiveLevels = map[string]interface{}{}
		l.ConfiguredLevels = map[string]interface{}{}
		// Note, logger name doesn't matter in "single-entry"
		if len(expectedLevels) > 0 {
			l.EffectiveLevels["logger"] = expectedLevels[0]
		}
		if len(expectedLevels) > 1 {
			l.ConfiguredLevels["logger"] = expectedLevels[1]
		}
	}
}

func ExpectLoggersEffectiveLevels(kvs...string) ExpectedLoggersOptions {
	return func(l *ExpectedLoggers) {
		setKVs(l.EffectiveLevels, kvs)
	}
}

func ExpectLoggersConfiguredLevels(kvs...string) ExpectedLoggersOptions {
	return func(l *ExpectedLoggers) {
		setKVs(l.ConfiguredLevels, kvs)
	}
}

func setKVs(kvMap map[string]interface{}, kvs []string) {
	for i := range kvs {
		var v string
		if i + 1 < len(kvs) {
			v = kvs[i+1]
		}
		kvMap[kvs[i]] = v
	}
}

// AssertLoggersResponse fail the test if the response is not the response of "loggers" endpoint.
// By default, this function expects:
//	- The response includes all loggers with effective level and all supported levels.
// This function only support V3 response.
func AssertLoggersResponse(t *testing.T, resp *http.Response, expectations ...ExpectedLoggersOptions) {
	expected := ExpectedLoggers{
		EffectiveLevels:  map[string]interface{}{},
		ConfiguredLevels: map[string]interface{}{},
	}
	for _, fn := range expectations {
		fn(&expected)
	}

	g := gomega.NewWithT(t)
	switch {
	case expected.Single:
		assertSingleLoggerResponse(t, g, resp, &expected)
	default:
		assertLoggersResponse(t, g, resp, &expected)
	}
}

func assertLoggersResponse(t *testing.T, g *WithT, resp *http.Response, expected *ExpectedLoggers) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `loggers response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.levels"), "loggers response should contains 'levels'")
	g.Expect(body).To(HaveJsonPath("$.loggers"), "loggers response should contains 'loggers'")
	g.Expect(body).To(HaveJsonPath("$.loggers[*].effectiveLevel"), "loggers response should contains 'effectiveLevel'")
	assertLogLevels(t, g, body, expected, func(name string) string {
		return fmt.Sprintf(`$.loggers["%s"]`, name)
	})
}

func assertSingleLoggerResponse(t *testing.T, g *WithT, resp *http.Response, expected *ExpectedLoggers) {
	body, e := io.ReadAll(resp.Body)
	g.Expect(e).To(Succeed(), `loggers response body should be readable`)
	g.Expect(body).To(HaveJsonPath("$.effectiveLevel"), "loggers response should contains 'effectiveLevel'")
	assertLogLevels(t, g, body, expected, func(_ string) string {
		return "$"
	})
}

func assertLogLevels(_ *testing.T, g *WithT, body []byte, expected *ExpectedLoggers, loggerJsonPathFn func(name string) string) {
	for k, v := range expected.EffectiveLevels {
		jsonPath := loggerJsonPathFn(k) + ".effectiveLevel"
		if v == nil {
			g.Expect(body).To(HaveJsonPath(jsonPath), "loggers response should contains logger '%s' with effectiveLevel", k)
		} else {
			g.Expect(body).To(HaveJsonPathWithValue(jsonPath, v), "loggers response should contains logger '%s' with effectiveLevel=%s", k, v)
		}
	}

	for k, v := range expected.ConfiguredLevels {
		jsonPath := loggerJsonPathFn(k) + ".configuredLevel"
		if v == nil {
			g.Expect(body).To(HaveJsonPath(jsonPath), "loggers response should contains logger '%s' with configuredLevel", k)
		} else {
			g.Expect(body).To(HaveJsonPathWithValue(jsonPath, v), "loggers response should contains logger '%s' with configuredLevel=%s", k, v)
		}
	}
}
