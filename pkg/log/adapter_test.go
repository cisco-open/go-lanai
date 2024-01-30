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
	"testing"
)

/*************************
	Tests
 *************************/

func TestWriterAdapter(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupApplyTestLoggerConfig()),
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.GomegaSubTest(SubTestWrite(LevelDebug), "DebugWithContext"),
		test.GomegaSubTest(SubTestWrite(LevelInfo), "InfoWithContext"),
		test.GomegaSubTest(SubTestWrite(LevelWarn), "WarnWithContext"),
		test.GomegaSubTest(SubTestWrite(LevelError), "ErrorWithContext"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestWrite(level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(`log/adapter_test\.go:[0-9]+`),
		)
		expectJson := CopyOf(expectText)

		logger := New(LoggerName)
		writer := NewWriterAdapter(logger, level)

		// With Context
		msg, _ := RandomMessage()
		n, e := writer.Write([]byte(msg))
		g.Expect(e).To(Succeed(), "Write() should not fail")
		g.Expect(n).ToNot(BeZero(), "Write() should return non-zero count")
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))
	}
}