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

/*************************
	Setup Test
 *************************/

const (
	LogOutputJson = `testdata/.tmp/logs/json.log`
	LogOutputText = `testdata/.tmp/logs/text.log`
)

const (
	StaticLoggedCtxKey   = `k-ctx-test`
	StaticKeyInLog       = `from-ctx`
	StaticLoggedCtxValue = `test-value-in-ctx`
)

/*************************
	Tests
 *************************/

func TestGoKitLogger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.SubTestSetup(SubSetupTestContext()),
		test.GomegaSubTest(SubTestLoggingWithContext(GoKitFactoryCreator(), LevelDebug), "DebugWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(GoKitFactoryCreator(), LevelDebug), "DebugWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(GoKitFactoryCreator(), LevelInfo), "InfoWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(GoKitFactoryCreator(), LevelInfo), "InfoWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(GoKitFactoryCreator(), LevelWarn), "WarnWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(GoKitFactoryCreator(), LevelWarn), "WarnWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(GoKitFactoryCreator(), LevelError), "ErrorWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(GoKitFactoryCreator(), LevelError), "ErrorWithoutContext"),
		test.GomegaSubTest(SubTestWithCaller(GoKitFactoryCreator(), LevelInfo), "WithCaller"),
		test.GomegaSubTest(SubTestTerminal(GoKitFactoryCreator()), "IsTerminal"),
	)
}

func TestZapLogger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.SubTestSetup(SubSetupTestContext()),
		test.GomegaSubTest(SubTestLoggingWithContext(ZapFactoryCreator(), LevelDebug), "DebugWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(ZapFactoryCreator(), LevelDebug), "DebugWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(ZapFactoryCreator(), LevelInfo), "InfoWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(ZapFactoryCreator(), LevelInfo), "InfoWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(ZapFactoryCreator(), LevelWarn), "WarnWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(ZapFactoryCreator(), LevelWarn), "WarnWithoutContext"),
		test.GomegaSubTest(SubTestLoggingWithContext(ZapFactoryCreator(), LevelError), "ErrorWithContext"),
		test.GomegaSubTest(SubTestLoggingWithoutContext(ZapFactoryCreator(), LevelError), "ErrorWithoutContext"),
		test.GomegaSubTest(SubTestWithCaller(ZapFactoryCreator(), LevelInfo), "WithCaller"),
		test.GomegaSubTest(SubTestTerminal(ZapFactoryCreator()), "IsTerminal"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubSetupClearLogOutput() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if e := os.WriteFile(LogOutputJson, []byte{}, 0666); e != nil {
			return ctx, e
		}
		if e := os.WriteFile(LogOutputText, []byte{}, 0666); e != nil {
			return ctx, e
		}
		return ctx, nil
	}
}

func SubSetupTestContext() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		ctx = context.WithValue(ctx, StaticLoggedCtxKey, StaticLoggedCtxValue)
		return ctx, nil
	}
}

func SubTestLoggingWithContext(fn TestFactoryCreateFunc, level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectFields(StaticKeyInLog, "test-value-in-ctx"),
		)
		expectJson := CopyOf(expectText)

		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)

		// With Context
		AssertLeveledLogging(g, logger.WithContext(ctx), level, expectJson, expectText)
	}
}

func SubTestLoggingWithoutContext(fn TestFactoryCreateFunc, level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectFields(StaticKeyInLog, nil),
		)
		expectJson := CopyOf(expectText)

		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)

		// Without Context
		AssertLeveledLogging(g, logger, level, expectJson, expectText)
	}
}

func SubTestWithCaller(fn TestFactoryCreateFunc, level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		const staticCaller = `SubTest`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(`SubTest`),
			ExpectFields(StaticKeyInLog, nil),
		)
		expectJson := CopyOf(expectText)

		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		AssertLeveledLogging(g, logger.WithCaller(staticCaller), level, expectJson, expectText)
	}
}

func SubTestTerminal(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		g.Expect(IsTerminal(logger)).To(BeFalse(), "IsTerminal should be false for multi-dest logger")
	}
}

/*************************
	Helpers
 *************************/

type TestFactoryCreateFunc func(g *gomega.WithT, fsys fs.FS, path string) loggerFactory

func GoKitFactoryCreator() TestFactoryCreateFunc {
	return func(g *WithT, fsys fs.FS, path string) loggerFactory {
		data, e := fs.ReadFile(fsys, path)
		g.Expect(e).To(Succeed(), "reading logger config file should not fail")
		var p Properties
		e = yaml.Unmarshal(data, &p)
		g.Expect(e).To(Succeed(), "parsing logger config file should not fail")
		return newKitLoggerFactory(&p)
	}
}

func ZapFactoryCreator() TestFactoryCreateFunc {
	return func(g *WithT, fsys fs.FS, path string) loggerFactory {
		data, e := fs.ReadFile(fsys, path)
		g.Expect(e).To(Succeed(), "reading logger config file should not fail")
		var p Properties
		e = yaml.Unmarshal(data, &p)
		g.Expect(e).To(Succeed(), "parsing logger config file should not fail")
		return newZapLoggerFactory(&p)
	}
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

type ExpectedLog struct {
	Name        string
	Level       LoggingLevel
	Msg         string
	CallerRegex string
	Fields      map[string]interface{}
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

func AssertLeveledLogging(g *gomega.WithT, ctxLogger Logger, level LoggingLevel, expectJson, expectText *ExpectedLog) {
	const StaticMsg = `test log with random token`
	var token string
	var msg string

	token = utils.RandomString(10)
	msg = fmt.Sprintf("%s [%s]", StaticMsg, token)
	DoLeveledLogf(ctxLogger, level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields(LogKeyMessage, msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	DoLeveledLogf(ctxLogger.WithKV("adhoc-key", "adhoc-value"),
		level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))

	DoLeveledLogKV(ctxLogger, level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value")))

	DoLeveledLogKV(ctxLogger.WithKV("adhoc-key", "adhoc-value"),
		level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))

	e := ctxLogger.WithLevel(level).Log(LogKeyMessage, msg)
	g.Expect(e).To(Succeed(), "Log() should not fail")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	ctxLogger.WithLevel(level).Printf("%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	ctxLogger.WithLevel(level).Print(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	ctxLogger.WithLevel(level).Println(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg+"\n")))
}

func AssertJsonLogEntry(g *gomega.WithT, logEntry string, expect *ExpectedLog) {
	g.Expect(logEntry).ToNot(BeEmpty(), "log should not be empty")
	var decoded map[string]interface{}
	e := json.Unmarshal([]byte(logEntry), &decoded)
	g.Expect(e).To(Succeed(), "parsing JSON log should not fail")

	g.Expect(decoded).To(HaveKeyWithValue(LogKeyLevel, strings.ToLower(expect.Level.String())), "log should have correct level")
	g.Expect(decoded).To(HaveKeyWithValue(LogKeyName, expect.Name), "log should have correct logger name")
	g.Expect(decoded).To(HaveKeyWithValue(LogKeyMessage, expect.Msg), "log should have correct message")
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

func AssertTextLogEntry(g *gomega.WithT, logEntry string, expect *ExpectedLog) {
	g.Expect(logEntry).ToNot(BeEmpty(), "log should not be empty")
	if len(expect.CallerRegex) == 0 {
		expect.CallerRegex = `*[a-zA-Z0-9_\-]+.go:[0-9]+`
	}

	regex := fmt.Sprintf(`[0-9\-]+T[0-9:.]+Z +%s +\[ *%s\] +%s: .*%s.*`,
		regexp.QuoteMeta(expect.Level.String()), expect.CallerRegex, regexp.QuoteMeta(expect.Name), regexp.QuoteMeta(expect.Msg))

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
