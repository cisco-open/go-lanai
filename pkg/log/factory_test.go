package log

import (
	"bufio"
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"encoding/json"
	"fmt"
	"github.com/ghodss/yaml"
	"github.com/go-kit/kit/log"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io"
	"io/fs"
	"os"
	"regexp"
	"strings"
	"sync"
	"testing"
	"time"
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
	StaticKeyInLogAlt    = `from-ctx-alt`
	StaticKeyInLogExtra  = `extra-key`
	StaticLoggedCtxValue = `test-value-in-ctx`
	StaticMsg            = `test log with random token`
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
		test.GomegaSubTest(SubTestWithCaller(GoKitFactoryCreator()), "WithCaller"),
		test.GomegaSubTest(SubTestConcurrent(GoKitFactoryCreator()), "Concurrent"),
		test.GomegaSubTest(SubTestRefresh(GoKitFactoryCreator()), "Refresh"),
		test.GomegaSubTest(SubTestAddContextValuers(GoKitFactoryCreator()), "AddContextValuers"),
		test.GomegaSubTest(SubTestSetLevel(GoKitFactoryCreator()), "SetLevel"),
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
		test.GomegaSubTest(SubTestWithCaller(ZapFactoryCreator()), "WithCaller"),
		test.GomegaSubTest(SubTestConcurrent(ZapFactoryCreator()), "Concurrent"),
		test.GomegaSubTest(SubTestRefresh(ZapFactoryCreator()), "TestRefresh"),
		test.GomegaSubTest(SubTestAddContextValuers(ZapFactoryCreator()), "AddContextValuers"),
		test.GomegaSubTest(SubTestSetLevel(ZapFactoryCreator()), "SetLevel"),
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

func SubTestWithCaller(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		const staticCaller = `SubTest`
		level := LevelInfo
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		// static caller
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(`SubTest`),
			ExpectFields(StaticKeyInLog, nil),
		)
		expectJson := CopyOf(expectText)
		valuer := func() interface{} { return staticCaller }
		AssertLeveledLogging(g, logger.WithCaller(staticCaller), level, expectJson, expectText)
		AssertLeveledLogging(g, logger.WithCaller(valuer), level, expectJson, expectText)

		// runtime caller with different skip depth
		expectText = CopyOf(expectText, ExpectCaller(`(testing|test|subtest)\.go:[0-9]+`))
		expectJson = CopyOf(expectJson, ExpectCaller(`(testing|test|subtest)\.go:[0-9]+`))
		switch f.(type) {
		case *kitLoggerFactory:
			AssertLeveledLogging(g, logger.WithCaller(log.Caller(7)), level, expectJson, expectText)
		case *zapLoggerFactory:
			AssertLeveledLogging(g, logger.WithCaller(RuntimeCaller(7)), level, expectJson, expectText)
		}
	}
}

func SubTestConcurrent(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		const routines = 500
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)

		// log many stuff concurrently
		var wg sync.WaitGroup
		wg.Add(routines)
		for i := 0; i < routines; i++ {
			go func() {
				token := utils.RandomString(10)
				logger.WithContext(ctx).WithKV(token, time.Second).Infof("%s [%s]", StaticMsg, token)
				wg.Done()
			}()
		}
		wg.Wait()

		// assert results
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(LevelInfo),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectMsgRegex(fmt.Sprintf(`%s \[%s\]`, regexp.QuoteMeta(StaticMsg), "[a-zA-Z0-9]{10}")),
		)
		expectJson := CopyOf(expectText)
		AssertEachJsonLogEntry(g, routines, expectJson)
		AssertEachTextLogEntry(g, routines, expectText)

	}
}

func SubTestRefresh(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(LevelInfo),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectFields(StaticKeyInLog, "test-value-in-ctx"),
		)
		expectJson := CopyOf(expectText)
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelInfo, expectJson, expectText)

		expectText = CopyOf(expectText, ExpectLevel(LevelWarn))
		expectJson = CopyOf(expectJson, ExpectLevel(LevelWarn))
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelWarn, expectJson, expectText)

		// refresh with different log level and context mappers
		newProps := BindProperties(g, os.DirFS("testdata"), "multi-dest-alt.yml")
		e := f.refresh(&newProps)
		g.Expect(e).To(Succeed(), "refresh should not fail")

		expectText = CopyOf(expectText, ExpectLevel(LevelWarn), ExpectFields(StaticKeyInLog, nil, StaticKeyInLogAlt, "test-value-in-ctx"))
		expectJson = CopyOf(expectJson, ExpectLevel(LevelWarn), ExpectFields(StaticKeyInLog, nil, StaticKeyInLogAlt, "test-value-in-ctx"))
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelWarn, expectJson, expectText)
		// logging at level INFO should not yield new logs (last log is still the same
		expectText = CopyOf(expectText, ExpectNotExists(), ExpectLevel(LevelInfo))
		expectJson = CopyOf(expectJson, ExpectNotExists(), ExpectLevel(LevelInfo))
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelInfo, expectJson, expectText)

		// refresh with different loggers
		newProps = BindProperties(g, os.DirFS("testdata"), "no-loggers.yml")
		e = f.refresh(&newProps)
		g.Expect(e).To(Succeed(), "refresh should not fail")
		expectText = CopyOf(expectText, ExpectNotExists(), ExpectLevel(LevelWarn))
		expectJson = CopyOf(expectJson, ExpectNotExists(), ExpectLevel(LevelWarn))
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelWarn, expectJson, expectText)
	}
}

func SubTestAddContextValuers(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName1 = `TestLogger`
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName1)
		expectText := NewExpectedLog(
			ExpectName(LoggerName1),
			ExpectLevel(LevelDebug),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectFields(StaticKeyInLog, "test-value-in-ctx"),
		)
		expectJson := CopyOf(expectText)
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelDebug, expectJson, expectText)

		// try add new context valuers
		f.addContextValuers(ContextValuers{
			StaticKeyInLogExtra: func(_ context.Context) interface{} {
				return "extra-value"
			},
		})
		expectText = CopyOf(expectText, ExpectLevel(LevelDebug), ExpectFields(StaticKeyInLogExtra, "extra-value"))
		expectJson = CopyOf(expectJson, ExpectLevel(LevelDebug), ExpectFields(StaticKeyInLogExtra, "extra-value"))
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelDebug, expectJson, expectText)
	}
}

func SubTestSetLevel(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerNamePrefix = `TestLogger`
		const LoggerName1 = `TestLogger.1`
		const LoggerName2 = `TestLogger.2`
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger1 := f.createLogger(LoggerName1)
		logger2 := f.createLogger(LoggerName2)
		expectText1 := NewExpectedLog(
			ExpectName(LoggerName1),
			ExpectLevel(LevelInfo),
			ExpectCaller(`factory_test\.go:[0-9]+`),
			ExpectFields(StaticKeyInLog, "test-value-in-ctx"),
		)
		expectText2 := CopyOf(expectText1, ExpectName(LoggerName2))
		expectJson1 := CopyOf(expectText1)
		expectJson2 := CopyOf(expectJson1, ExpectName(LoggerName2))

		// set logger particular levels
		lvl := LevelInfo
		f.setLevel(LoggerName1, &lvl)
		f.setLevel(LoggerName2, &lvl)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelInfo, expectJson1, expectText1)
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelInfo, expectJson2, expectText2)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelDebug,
			CopyOf(expectJson1, ExpectNotExists(), ExpectLevel(LevelDebug)), CopyOf(expectText1, ExpectNotExists(), ExpectLevel(LevelDebug)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelDebug,
			CopyOf(expectJson2, ExpectNotExists(), ExpectLevel(LevelDebug)), CopyOf(expectText2, ExpectNotExists(), ExpectLevel(LevelDebug)))

		// try unset level, DEBUG should be enabled
		f.setLevel(LoggerName2, nil)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelDebug,
			CopyOf(expectJson1, ExpectNotExists(), ExpectLevel(LevelDebug)), CopyOf(expectText1,  ExpectNotExists(), ExpectLevel(LevelDebug)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelDebug,
			CopyOf(expectJson2, ExpectLevel(LevelDebug)), CopyOf(expectText2, ExpectLevel(LevelDebug)))

		// try set level with prefix, now logger 1 has specific settings, logger 2 inherits parent logger settings
		lvl = LevelWarn
		f.setLevel(LoggerNamePrefix, &lvl)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelInfo,
			CopyOf(expectJson1, ExpectLevel(LevelInfo)), CopyOf(expectText1,  ExpectLevel(LevelInfo)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelInfo,
			CopyOf(expectJson2, ExpectNotExists(), ExpectLevel(LevelInfo)), CopyOf(expectText2, ExpectNotExists(), ExpectLevel(LevelInfo)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelWarn,
			CopyOf(expectJson2, ExpectLevel(LevelWarn)), CopyOf(expectText2, ExpectLevel(LevelWarn)))

		// try set root and unset everything else
		lvl = LevelError
		f.setLevel("default", &lvl)
		f.setLevel(LoggerNamePrefix, nil)
		f.setLevel(LoggerName1, nil)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelWarn,
			CopyOf(expectJson1, ExpectNotExists(), ExpectLevel(LevelWarn)), CopyOf(expectText1, ExpectNotExists(), ExpectLevel(LevelWarn)))
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelError,
			CopyOf(expectJson1, ExpectLevel(LevelError)), CopyOf(expectText1,  ExpectLevel(LevelError)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelWarn,
			CopyOf(expectJson2, ExpectNotExists(), ExpectLevel(LevelWarn)), CopyOf(expectText2, ExpectNotExists(), ExpectLevel(LevelWarn)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelError,
			CopyOf(expectJson2, ExpectLevel(LevelError)), CopyOf(expectText2, ExpectLevel(LevelError)))
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

func BindProperties(g *gomega.WithT, fsys fs.FS, path string) (props Properties) {
	data, e := fs.ReadFile(fsys, path)
	g.Expect(e).To(Succeed(), "reading logger config file should not fail")
	e = yaml.Unmarshal(data, &props)
	g.Expect(e).To(Succeed(), "parsing logger config file should not fail")
	return
}

func GoKitFactoryCreator() TestFactoryCreateFunc {
	return func(g *WithT, fsys fs.FS, path string) loggerFactory {
		p := BindProperties(g, fsys, path)
		return newKitLoggerFactory(&p)
	}
}

func ZapFactoryCreator() TestFactoryCreateFunc {
	return func(g *WithT, fsys fs.FS, path string) loggerFactory {
		p := BindProperties(g, fsys, path)
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

func AssertLeveledLogging(g *gomega.WithT, l Logger, level LoggingLevel, expectJson, expectText *ExpectedLog) {
	var token string
	var msg string

	token = utils.RandomString(10)
	msg = fmt.Sprintf("%s [%s]", StaticMsg, token)
	DoLeveledLogf(l, level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields(LogKeyMessage, msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	DoLeveledLogf(l.WithKV("adhoc-key", "adhoc-value"),
		level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))

	DoLeveledLogKV(l, level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value")))

	DoLeveledLogKV(l.WithKV("adhoc-key", "adhoc-value"),
		level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))

	e := l.WithLevel(level).Log(LogKeyMessage, msg)
	g.Expect(e).To(Succeed(), "Log() should not fail")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	l.WithLevel(level).Printf("%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	l.WithLevel(level).Print(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	l.WithLevel(level).Println(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg+"\n")))
}

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
