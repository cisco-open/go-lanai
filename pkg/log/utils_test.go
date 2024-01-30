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
	"bufio"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"testing"
)

const (
	LogOutputJson = `testdata/.tmp/logs/json.log`
	LogOutputText = `testdata/.tmp/logs/text.log`
)

const (
	StaticMsg       = `test log with random token`
	CtxKeyStatic    = `k-ctx-test`
	CtxValueStatic  = `test-value-in-ctx`
	LogKeyStatic    = `from-ctx`
	LogKeyStaticAlt = `from-ctx-alt`
	LogKeyExtra     = `extra-key`
	LogKeySpanID    = `spanId`
	LogKeyTraceID   = `traceId`
	LogKeyParentID  = `parentId`
)

/*************************
	Tests
 *************************/

func TestUtilities(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestIsTerminal(), "IsTerminal"),
		test.GomegaSubTest(SubTestCapped(), "Capped"),
		test.GomegaSubTest(SubTestPadding(), "Padding"),
		test.GomegaSubTest(SubTestDebugShowcase(), "DebugShowcase"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestIsTerminal() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		logger := New(`TestLogger`)
		g.Expect(IsTerminal(logger)).To(BeFalse(), "IsTerminal should be false for any logger created in test")
	}
}

func SubTestCapped() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const ShortString = `short`
		const LongString = `this is very very long string and definitely should be truncated`
		const toCap = 10
		const ExpectedCappedTailing = `this is...`
		const ExpectedCappedMiddle = `thi...ated`
		var capped string
		// tailing
		capped = Capped(ShortString, toCap)
		g.Expect(capped).To(Equal(ShortString), "Capped() with short string and positive number should not change")
		capped = Capped(LongString, toCap)
		g.Expect(capped).To(Equal(ExpectedCappedTailing), "Capped() with string and positive number should have correct result")

		// middle
		capped = Capped(ShortString, -toCap)
		g.Expect(capped).To(Equal(ShortString), "Capped() with short string and negative number should not change")
		capped = Capped(LongString, -toCap)
		g.Expect(capped).To(Equal(ExpectedCappedMiddle), "Capped() with string and negative number should have correct result")

		// zero cap
		capped = Capped(LongString, 0)
		g.Expect(capped).To(BeEmpty(), "Capped() with string and zero number should have correct result")
	}
}

func SubTestPadding() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const Word = `word`
		const padding = 10
		const ExpectedLeftPadding = `      word`
		const ExpectedRightPadding = `word      `
		var padded string
		// left
		padded = Padding(Word, padding)
		g.Expect(padded).To(Equal(ExpectedLeftPadding), "Padding() with string and positive number should have correct result")

		// right
		padded = Padding(Word, -padding)
		g.Expect(padded).To(Equal(ExpectedRightPadding), "Padding() with string and negative number should have correct result")

		// zero cap
		padded = Padding(Word, 0)
		g.Expect(padded).To(Equal(Word), "Padding() with string and zero number should have correct result")
	}
}

func SubTestDebugShowcase() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		DebugShowcase()
		// nothing we could verify here, as long as it doesn't panic
		// colored output should be seen in console (if applicable)
	}
}

/*************************
	Helpers
 *************************/

type TestFactoryCreateFunc func(g *gomega.WithT, fsys fs.FS, path string) loggerFactory

func BindProperties(g *gomega.WithT, fsys fs.FS, path string) (props Properties) {
	data, e := fs.ReadFile(fsys, path)
	g.Expect(e).To(Succeed(), "reading logger config file should not fail")
	e = yaml.Unmarshal(data, &props)
	g.Expect(e).To(Succeed(), "parsing logger config file should not fail")
	return
}

func RandomMessage() (msg, token string) {
	token = utils.RandomString(10)
	msg = fmt.Sprintf("%s [%s]", StaticMsg, token)
	return msg, token
}

func DoLeveledLogf(l Logger, lvl LoggingLevel, tmpl string, args ...interface{}) {
	switch lvl {
	case LevelDebug:
		l.Debugf(tmpl, args...)
	case LevelInfo:
		l.Infof(tmpl, args...)
	case LevelWarn:
		l.Warnf(tmpl, args...)
	case LevelError:
		l.Errorf(tmpl, args...)
	default:
		// do nothing
	}
}

func DoLeveledLogKV(l Logger, lvl LoggingLevel, msg string, kvs ...interface{}) {
	switch lvl {
	case LevelDebug:
		l.Debug(msg, kvs...)
	case LevelInfo:
		l.Info(msg, kvs...)
	case LevelWarn:
		l.Warn(msg, kvs...)
	case LevelError:
		l.Error(msg, kvs...)
	default:
		// do nothing
	}
}

/*************************
	Expectation
 *************************/

type ExpectedLog struct {
	Name        string
	Level       LoggingLevel
	Msg         string
	MsgRegex    string
	CallerRegex string
	Fields      map[string]interface{}
	Absent      bool
}

func NewExpectedLog(opts ...func(expect *ExpectedLog)) *ExpectedLog {
	expect := ExpectedLog{
		CallerRegex: `*[a-zA-Z0-9_\-]+.go:[0-9]+`,
		Fields:      map[string]interface{}{},
	}
	for _, fn := range opts {
		fn(&expect)
	}
	return &expect
}

func CopyOf(log *ExpectedLog, opts ...func(expect *ExpectedLog)) *ExpectedLog {
	if log == nil {
		return nil
	}
	cpy := *log
	cpy.Fields = map[string]interface{}{}
	for k, v := range log.Fields {
		cpy.Fields[k] = v
	}
	for _, fn := range opts {
		fn(&cpy)
	}
	return &cpy
}

func ExpectNotExists() func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.Absent = true
	}
}

func ExpectName(name string) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.Name = name
	}
}

func ExpectLevel(lvl LoggingLevel) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.Level = lvl
	}
}

func ExpectMsg(msg string) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.Msg = msg
	}
}

func ExpectMsgRegex(regex string) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.MsgRegex = regex
	}
}

func ExpectCaller(callerRegex string) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.CallerRegex = callerRegex
	}
}

func ExpectFields(kvs ...interface{}) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		for i := 0; i < len(kvs); i += 2 {
			var key string
			switch k := kvs[i].(type) {
			case string:
				key = k
			default:
				key = fmt.Sprint(k)
			}
			if i == len(kvs)-1 {
				expect.Fields[key] = nil
			} else {
				expect.Fields[key] = kvs[i+1]
			}
		}
	}
}

/*************************
	Assertion
 *************************/

func AssertJsonLogEntry(g *gomega.WithT, logEntry string, expect *ExpectedLog) {
	if !expect.Absent {
		g.Expect(logEntry).ToNot(BeEmpty(), "log should not be empty")
	}
	var decoded map[string]interface{}
	e := json.Unmarshal([]byte(logEntry), &decoded)
	g.Expect(e).To(Succeed(), "parsing JSON log should not fail")

	if expect.Absent {
		if len(expect.MsgRegex) != 0 {
			g.Expect(decoded).ToNot(HaveKeyWithValue(LogKeyMessage, MatchRegexp(expect.MsgRegex)), "log should have correct message")
		} else {
			g.Expect(decoded).ToNot(HaveKeyWithValue(LogKeyMessage, expect.Msg), "log should have correct message")
		}
		return
	}
	g.Expect(decoded).To(HaveKeyWithValue(LogKeyLevel, strings.ToLower(expect.Level.String())), "log should have correct level")
	g.Expect(decoded).To(HaveKeyWithValue(LogKeyName, expect.Name), "log should have correct logger name")
	if len(expect.MsgRegex) != 0 {
		g.Expect(decoded).To(HaveKeyWithValue(LogKeyMessage, MatchRegexp(expect.MsgRegex)), "log should have correct message")
	} else {
		g.Expect(decoded).To(HaveKeyWithValue(LogKeyMessage, expect.Msg), "log should have correct message")
	}
	if len(expect.CallerRegex) != 0 {
		g.Expect(decoded).To(HaveKeyWithValue(LogKeyCaller, MatchRegexp(expect.CallerRegex)), "log should have correct caller")
	}
	for k, v := range expect.Fields {
		if v != nil {
			g.Expect(decoded).To(HaveKeyWithValue(k, v), "JSON log should have correct KV pair")
		} else {
			g.Expect(decoded).ToNot(HaveKey(k), "JSON log should not have key [%s]", k)
		}
	}
}

func AssertLastJsonLogEntry(g *gomega.WithT, expect *ExpectedLog) {
	line := ReadLastLogEntry(g, LogOutputJson, 0)
	AssertJsonLogEntry(g, string(line), expect)
}

func AssertEachJsonLogEntry(g *gomega.WithT, expectCount int, expect *ExpectedLog) {
	count := ForEachLogEntry(g, LogOutputJson, func(line []byte) {
		AssertJsonLogEntry(g, string(line), expect)
	})
	g.Expect(count).To(Equal(expectCount), "log should contain correct number of entries")

}

func AssertTextLogEntry(g *gomega.WithT, logEntry string, expect *ExpectedLog) {
	if !expect.Absent {
		g.Expect(logEntry).ToNot(BeEmpty(), "log should not be empty")
	}
	if len(expect.CallerRegex) == 0 {
		expect.CallerRegex = `*[a-zA-Z0-9_\-]+.go:[0-9]+`
	}
	msgRegex := expect.MsgRegex
	if len(msgRegex) == 0 {
		msgRegex = regexp.QuoteMeta(expect.Msg)
	}
	regex := fmt.Sprintf(`[0-9\-]+T[0-9:.]+Z +%s +\[ *%s\] +%s: .*%s.*`,
		regexp.QuoteMeta(expect.Level.String()), expect.CallerRegex, regexp.QuoteMeta(expect.Name), msgRegex)

	if expect.Absent {
		g.Expect(logEntry).ToNot(MatchRegexp(regex), "log should not match pattern")
		return
	}
	g.Expect(logEntry).To(MatchRegexp(regex), "log should have correct pattern")
	for k, v := range expect.Fields {
		switch p := v.(type) {
		case nil:
			regex := fmt.Sprintf(`%s *= *[^ ]+`, regexp.QuoteMeta(k))
			g.Expect(logEntry).ToNot(MatchRegexp(regex), "log should not have field [%s]", k)
		case regexp.Regexp:
			g.Expect(logEntry).To(MatchRegexp(p.String()), "log should have correct fields")
		default:
			regex := fmt.Sprintf(`%s *= *"%s"`, regexp.QuoteMeta(k), regexp.QuoteMeta(fmt.Sprint(v)))
			g.Expect(logEntry).To(MatchRegexp(regex), "log should have correct fields")
		}
	}
}

func AssertLastTextLogEntry(g *gomega.WithT, expect *ExpectedLog) {
	line := ReadLastLogEntry(g, LogOutputText, 0)
	AssertTextLogEntry(g, string(line), expect)
}

func AssertEachTextLogEntry(g *gomega.WithT, expectCount int, expect *ExpectedLog) {
	count := ForEachLogEntry(g, LogOutputText, func(line []byte) {
		AssertTextLogEntry(g, string(line), expect)
	})
	g.Expect(count).To(Equal(expectCount), "log should contain correct number of entries")
}

func ForEachLogEntry(g *gomega.WithT, logPath string, fn func([]byte)) (count int) {
	file, e := os.Open(logPath)
	g.Expect(e).To(Succeed(), "reading log output [%s] should not fail", logPath)
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	// read first non-empty line and use it to estimate last line's length
	for scanner.Scan() {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		fn(scanner.Bytes())
		count++
	}
	return count
}

func ReadLastLogEntry(g *gomega.WithT, logPath string, maxLineLen int64) (lastLine []byte) {
	file, e := os.Open(logPath)
	g.Expect(e).To(Succeed(), "reading log output [%s] should not fail", logPath)
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	// read first non-empty line and use it to estimate last line's length
	for scanner.Scan() {
		if len(scanner.Bytes()) != 0 {
			lastLine = scanner.Bytes()
			break
		}
	}
	if len(lastLine) == 0 {
		return
	} else if maxLineLen < int64(len(lastLine)*3) {
		maxLineLen = int64(len(lastLine) * 3)
	}

	// set offset
	stat, e := file.Stat()
	g.Expect(e).To(Succeed(), "getting stat of log output [%s] should not fail", logPath)
	if stat.Size() > maxLineLen {
		offset, e := file.Seek(-maxLineLen, io.SeekEnd)
		g.Expect(e).To(Succeed(), "setting read offset to %d should not fail", maxLineLen)
		g.Expect(offset).To(BeNumerically(">=", 0), "seeking offset [%d] should be positive", offset)
	}

	// continue scan near the end
	for scanner.Scan() {
		if len(scanner.Bytes()) != 0 {
			lastLine = scanner.Bytes()
		}
	}
	return lastLine
}
