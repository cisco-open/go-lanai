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
	. "github.com/onsi/gomega"
	"testing"
)

/*************************
	Tests
 *************************/

func TestManageLogger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupApplyTestLoggerConfig()),
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.GomegaSubTest(SubTestManageGetLevel(), "Levels"),
		test.GomegaSubTest(SubTestManageSetLevel(), "SetLevel"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestManageGetLevel() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerKeyPrefix = `test-logger`
		const LoggerNamePrefix = `TestLogger`
		const LoggerName1 = `TestLogger.1`
		const LoggerKey1 = `test-logger.1`
		const LoggerName2 = `TestLogger.2`
		const LoggerKey2 = `test-logger.2`
		_ = New(LoggerName1)
		_ = New(LoggerName2)

		rs := Levels("")
		g.Expect(rs).To(HaveLen(4), "Levels() with empty prefix should have correct length")
		AssertLevelConfigs(g, rs, "default", "ROOT", LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelDebug, -1)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelDebug, -1)

		rs = Levels(LoggerNamePrefix)
		g.Expect(rs).To(HaveLen(3), "Levels() with prefix should have correct length")
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelDebug, -1)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelDebug, -1)
	}
}

func SubTestManageSetLevel() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerKeyPrefix = `test-logger`
		const LoggerNamePrefix = `TestLogger`
		const LoggerName1 = `TestLogger.1`
		const LoggerKey1 = `test-logger.1`
		const LoggerName2 = `TestLogger.2`
		const LoggerKey2 = `test-logger.2`
		var rs map[string]*LevelConfig

		_ = New(LoggerName1)
		_ = New(LoggerName2)

		// set logger particular levels
		lvl := LevelInfo
		SetLevel(LoggerName1, &lvl)
		SetLevel(LoggerName2, &lvl)
		rs = Levels("")
		AssertLevelConfigs(g, rs, "default", "ROOT", LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelInfo, LevelInfo)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelInfo, LevelInfo)

		// try unset level, DEBUG should be enabled
		SetLevel(LoggerName2, nil)
		rs = Levels("")
		AssertLevelConfigs(g, rs, "default", "ROOT", LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelInfo, LevelInfo)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelDebug, -1)

		// try set level with prefix, now logger 1 has specific settings, logger 2 inherits parent logger settings
		lvl = LevelWarn
		SetLevel(LoggerNamePrefix, &lvl)
		rs = Levels("")
		AssertLevelConfigs(g, rs, "default", "ROOT", LevelDebug, LevelDebug)
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelWarn, LevelWarn)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelInfo, LevelInfo)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelWarn, -1)

		// try set root and unset everything else
		lvl = LevelError
		SetLevel("default", &lvl)
		SetLevel(LoggerNamePrefix, nil)
		SetLevel(LoggerName1, nil)
		rs = Levels("")
		AssertLevelConfigs(g, rs, "default", "ROOT", LevelError, LevelError)
		AssertLevelConfigs(g, rs, LoggerKeyPrefix, LoggerNamePrefix, LevelError, -1)
		AssertLevelConfigs(g, rs, LoggerKey1, LoggerName1, LevelError, -1)
		AssertLevelConfigs(g, rs, LoggerKey2, LoggerName2, LevelError, -1)
	}
}

/*************************
	Helpers
 *************************/

func AssertLevelConfig(g *gomega.WithT, cfg *LevelConfig, expectedName string, expectedEffective, expectedConfigured LoggingLevel) {
	g.Expect(cfg).ToNot(BeNil(), "level config should not be nil")
	g.Expect(cfg.Name).To(Equal(expectedName), "level config should have correct name")
	if expectedEffective >= 0 {
		g.Expect(cfg.EffectiveLevel).ToNot(BeNil(), "level config should have non-nil effective level")
		g.Expect(*cfg.EffectiveLevel).To(Equal(expectedEffective), "level config should have correct effective level")
	} else {
		g.Expect(cfg.EffectiveLevel).To(BeNil(), "level config should not have effective level")
	}
	if expectedConfigured >= 0 {
		g.Expect(cfg.ConfiguredLevel).ToNot(BeNil(), "level config should have non-nil effective level")
		g.Expect(*cfg.ConfiguredLevel).To(Equal(expectedConfigured), "level config should have correct effective level")
	} else {
		g.Expect(cfg.ConfiguredLevel).To(BeNil(), "level config should not have effective level")
	}
}

func AssertLevelConfigs(g *gomega.WithT, cfgs map[string]*LevelConfig, expectedKey, expectedName string, expectedEffective, expectedConfigured LoggingLevel) {
	if len(expectedKey) == 0 {
		g.Expect(cfgs).ToNot(HaveKey(expectedKey), "level configs should not have key")
		return
	}
	g.Expect(cfgs).To(HaveKey(expectedKey), "level configs should have key")
	cfg := cfgs[expectedKey]
	AssertLevelConfig(g, cfg, expectedName, expectedEffective, expectedConfigured)
}