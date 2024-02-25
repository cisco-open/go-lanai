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
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	"io/fs"
	"testing"
)

/*************************
	Tests
 *************************/

func TestZapLogger(t *testing.T) {
	factoryCreator := ZapFactoryCreator()
	test.RunTest(context.Background(), t,
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.SubTestSetup(SubSetupTestContext()),
		test.GomegaSubTest(SubTestLoggingWithContext(factoryCreator, LevelDebug), "DebugWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(factoryCreator, LevelDebug), "DebugWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(factoryCreator, LevelInfo), "InfoWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(factoryCreator, LevelInfo), "InfoWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(factoryCreator, LevelWarn), "WarnWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(factoryCreator, LevelWarn), "WarnWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(factoryCreator, LevelError), "ErrorWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(factoryCreator, LevelError), "ErrorWithoutContext"),
		test.GomegaSubTest(SubTestWithCaller(factoryCreator, RuntimeCaller(7)), "WithCaller"),
		test.GomegaSubTest(SubTestConcurrent(factoryCreator), "Concurrent"),
		test.GomegaSubTest(SubTestRefresh(factoryCreator), "TestRefresh"),
		test.GomegaSubTest(SubTestAddContextValuers(factoryCreator), "AddContextValuers"),
		test.GomegaSubTest(SubTestSetLevel(factoryCreator), "SetLevel"),
		test.GomegaSubTest(SubTestTerminal(factoryCreator), "IsTerminal"),
	)
}

/*************************
	Sub-Tests
 *************************/

/* Sub tests are defined in factory_test.go, shared with all factory tests */

/*************************
	Helpers
 *************************/

func ZapFactoryCreator() TestFactoryCreateFunc {
	return func(g *gomega.WithT, fsys fs.FS, path string) loggerFactory {
		p := BindProperties(g, fsys, path)
		return newZapLoggerFactory(&p)
	}
}
