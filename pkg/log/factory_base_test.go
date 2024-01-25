package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"errors"
	"fmt"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"
)

/*
	This file only defines sub tests that applicable to all factories implementation
*/

const ExpectedDefaultCaller = `utils_test\.go:[0-9]+`
const ExpectedDirectCaller = `factory_base_test\.go:[0-9]+`

/*************************
	Sub-Test Cases
 *************************/

func SubSetupClearLogOutput() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		if e := os.WriteFile(LogOutputJson, []byte{}, 0666); e != nil && !errors.Is(e, os.ErrNotExist) {
			return ctx, e
		}
		if e := os.WriteFile(LogOutputText, []byte{}, 0666); e != nil && !errors.Is(e, os.ErrNotExist) {
			return ctx, e
		}
		return ctx, nil
	}
}

func SubSetupTestContext() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		ctx = context.WithValue(ctx, CtxKeyStatic, CtxValueStatic)
		return ctx, nil
	}
}

func SubTestLoggingWithContext(fn TestFactoryCreateFunc, level LoggingLevel) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(level),
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, "test-value-in-ctx"),
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
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, nil),
		)
		expectJson := CopyOf(expectText)

		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)

		// Without Context
		AssertLeveledLogging(g, logger, level, expectJson, expectText)
	}
}

func SubTestWithCaller(fn TestFactoryCreateFunc, runtimeCaller interface{}) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		const staticCaller = `SubTest`
		var msg string
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		// static caller
		expectText := NewExpectedLog(
			ExpectName(LoggerName),
			ExpectLevel(LevelInfo),
			ExpectCaller(`SubTest`),
			ExpectFields(LogKeyStatic, nil),
		)
		expectJson := CopyOf(expectText)
		valuer := func() interface{} { return staticCaller }

		msg, _ = RandomMessage()
		logger.WithCaller(staticCaller).Info(msg)
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

		msg, _ = RandomMessage()
		logger.WithCaller(valuer).Info(msg)
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

		// runtime caller with different skip depth
		msg, _ = RandomMessage()
		expectText = CopyOf(expectText, ExpectCaller(`(testing|test|subtest)\.go:[0-9]+`))
		expectJson = CopyOf(expectJson, ExpectCaller(`(testing|test|subtest)\.go:[0-9]+`))
		logger.WithCaller(runtimeCaller).Info(msg)
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg)))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))
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
			ExpectCaller(ExpectedDirectCaller),
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
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, "test-value-in-ctx"),
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

		expectText = CopyOf(expectText, ExpectLevel(LevelWarn), ExpectFields(LogKeyStatic, nil, LogKeyStaticAlt, "test-value-in-ctx"))
		expectJson = CopyOf(expectJson, ExpectLevel(LevelWarn), ExpectFields(LogKeyStatic, nil, LogKeyStaticAlt, "test-value-in-ctx"))
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
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, "test-value-in-ctx"),
		)
		expectJson := CopyOf(expectText)
		AssertLeveledLogging(g, logger.WithContext(ctx), LevelDebug, expectJson, expectText)

		// try add new context valuers
		f.addContextValuers(ContextValuers{
			LogKeyExtra: func(_ context.Context) interface{} {
				return "extra-value"
			},
		})
		expectText = CopyOf(expectText, ExpectLevel(LevelDebug), ExpectFields(LogKeyExtra, "extra-value"))
		expectJson = CopyOf(expectJson, ExpectLevel(LevelDebug), ExpectFields(LogKeyExtra, "extra-value"))
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
			ExpectCaller(ExpectedDefaultCaller),
			ExpectFields(LogKeyStatic, "test-value-in-ctx"),
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
			CopyOf(expectJson1, ExpectNotExists(), ExpectLevel(LevelDebug)), CopyOf(expectText1, ExpectNotExists(), ExpectLevel(LevelDebug)))
		AssertLeveledLogging(g, logger2.WithContext(ctx), LevelDebug,
			CopyOf(expectJson2, ExpectLevel(LevelDebug)), CopyOf(expectText2, ExpectLevel(LevelDebug)))

		// try set level with prefix, now logger 1 has specific settings, logger 2 inherits parent logger settings
		lvl = LevelWarn
		f.setLevel(LoggerNamePrefix, &lvl)
		AssertLeveledLogging(g, logger1.WithContext(ctx), LevelInfo,
			CopyOf(expectJson1, ExpectLevel(LevelInfo)), CopyOf(expectText1, ExpectLevel(LevelInfo)))
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
			CopyOf(expectJson1, ExpectLevel(LevelError)), CopyOf(expectText1, ExpectLevel(LevelError)))
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

func AssertLeveledLogging(g *gomega.WithT, l Logger, level LoggingLevel, expectJson, expectText *ExpectedLog) {
	var token string
	var msg string

	msg, token = RandomMessage()
	DoLeveledLogf(l, level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields(LogKeyMessage, msg)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg)))

	msg, token = RandomMessage()
	DoLeveledLogf(l.WithKV("adhoc-key", "adhoc-value"),
		level, "%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("adhoc-key", "adhoc-value")))

	msg, token = RandomMessage()
	DoLeveledLogKV(l, level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value")))

	msg, token = RandomMessage()
	DoLeveledLogKV(l.WithKV("adhoc-key", "adhoc-value"),
		level, msg, "test-key", "test-value")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))

	// following logging statements has one less caller stack, therefore the "caller" would be different
	msg, token = RandomMessage()
	e := l.WithLevel(level).Log(LogKeyMessage, msg)
	g.Expect(e).To(Succeed(), "Log() should not fail")
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))

	msg, token = RandomMessage()
	l.WithLevel(level).Printf("%s [%s]", StaticMsg, token)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))

	msg, token = RandomMessage()
	l.WithLevel(level).Print(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))
	AssertLastTextLogEntry(g, CopyOf(expectText, ExpectMsg(msg), ExpectCaller(ExpectedDirectCaller)))

	msg, token = RandomMessage()
	l.WithLevel(level).Println(msg)
	AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectMsg(msg+"\n"), ExpectCaller(ExpectedDirectCaller)))
	// Note: multi-lined test format is not asserted
}
