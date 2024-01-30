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

package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"os"
	"testing"
)

/*************************
	Tests
 *************************/

func TestDefaultLogger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupApplyTestLoggerConfig()),
		test.Setup(SetupRegisterContextValuers()),
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.SubTestSetup(SubSetupTestContext()),
		test.GomegaSubTest(SubTestLoggerWithNew(LevelDebug), "DebugWithContext"),
		test.GomegaSubTest(SubTestLoggerWithNew(LevelInfo), "InfoWithContext"),
		test.GomegaSubTest(SubTestLoggerWithNew(LevelWarn), "WarnWithContext"),
		test.GomegaSubTest(SubTestLoggerWithNew(LevelError), "ErrorWithContext"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupApplyTestLoggerConfig() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		g := gomega.NewWithT(t)
		p := BindProperties(g, os.DirFS("testdata"), "multi-dest.yml")
		e := UpdateLoggingConfiguration(&p)
		g.Expect(e).To(Succeed(), "update default factory with configured properties should not fail")
		return ctx, nil
	}
}

func SetupRegisterContextValuers() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		RegisterContextLogFields(ContextValuers{
			LogKeyTraceID: func(ctx context.Context) interface{} {
				return "test-trace-id"
			},
			LogKeySpanID: func(ctx context.Context) interface{} {
				return "test-span-id"
			},
			LogKeyParentID: func(ctx context.Context) interface{} {
				return "test-trace-id"
			},
		})
		return ctx, nil
	}
}

func SubTestLoggerWithNew(level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, "test-value-in-ctx"),
		)
		expectJson := CopyOf(expectText, ExpectFields(
			LogKeyTraceID, "test-trace-id",
			LogKeySpanID, "test-span-id",
			LogKeyParentID, "test-trace-id",
		))

		logger := New(LoggerName)

		// With Context
		AssertLeveledLogging(g, logger.WithContext(ctx), level, expectJson, expectText)
	}
}

/*************************
	Helpers
 *************************/


