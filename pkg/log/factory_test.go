package log

import (
	"bufio"
	"context"
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
	StaticLoggedCtxKey = `k-ctx-test`
	StaticKeyInLog = `from-ctx`
	StaticLoggedCtxValue = `test-value-in-ctx`
)

/*************************
	Tests
 *************************/

func TestGoKitLogger(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.SubTestSetup(SubSetupClearLogOutput()),
		test.SubTestSetup(SubSetupTestContext()),
		test.GomegaSubTest(SubTestDebug(GoKitFactoryCreator()), "TestDebug"),
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

func SubTestDebug(fn TestFactoryCreateFunc) test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const LoggerName = `TestLogger`
		f := fn(g, os.DirFS("testdata"), "multi-dest.yml")
		logger := f.createLogger(LoggerName)
		expectJson := NewExpectedLog(ExpectFields(
			"logger", LoggerName,
			"level", "debug",
			StaticKeyInLog, "test-value-in-ctx",
		))
		expectText := NewExpectedLog(
			ExpectTextRegex(LoggerName, LevelDebug, "test-log"),
			ExpectFields(
				StaticKeyInLog, "test-value-in-ctx",
			),
		)

		logger.WithContext(ctx).Debugf("test-log")
		AssertLastJsonLogEntry(g, expectJson)
		AssertLastTextLogEntry(g, expectText)
		logger.WithContext(ctx).Debug("test-log", "test-key", "test-value")
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectFields("test-key", "test-value")))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectFields("test-key", "test-value")))
		logger.WithContext(ctx).WithKV("adhoc-key", "adhoc-value").Debug("test-log", "test-key", "test-value")
		AssertLastJsonLogEntry(g, CopyOf(expectJson, ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))
		AssertLastTextLogEntry(g, CopyOf(expectText, ExpectFields("test-key", "test-value", "adhoc-key", "adhoc-value")))
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

type ExpectedLog struct {
	Regex  string
	Fields map[string]interface{}
}

func NewExpectedLog(opts ...func(expect *ExpectedLog)) *ExpectedLog {
	expect := ExpectedLog{
		Fields: map[string]interface{}{},
	}
	for _, fn := range opts {
		fn(&expect)
	}
	return &expect
}

func CopyOf(log *ExpectedLog, opts ...func(expect *ExpectedLog)) *ExpectedLog {
	expect := *log
	for _, fn := range opts {
		fn(&expect)
	}
	return &expect
}

func ExpectRegex(regex string) func(expect *ExpectedLog) {
	return func(expect *ExpectedLog) {
		expect.Regex = regex
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

func ExpectTextRegex(loggerName string, level LoggingLevel, msg string) func(expect *ExpectedLog) {
	regex := fmt.Sprintf(`[0-9\-]+T[0-9:.]+Z +%s +\[ *[a-zA-Z0-9_\-]+.go:[0-9]+\] +%s: .*%s.*`, level.String(), loggerName, msg)
	return ExpectRegex(regex)
}

func AssertJsonLogEntry(g *gomega.WithT, logEntry string, expect *ExpectedLog) {
	g.Expect(logEntry).ToNot(BeEmpty(), "log should not be empty")
	if len(expect.Regex) != 0 {
		g.Expect(logEntry).To(MatchRegexp(expect.Regex), "log should have correct pattern")
	}
	var decoded map[string]interface{}
	e := json.Unmarshal([]byte(logEntry), &decoded)
	g.Expect(e).To(Succeed(), "parsing JSON log should not fail")
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
	if len(expect.Regex) != 0 {
		g.Expect(logEntry).To(MatchRegexp(expect.Regex), "log should have correct pattern")
	}
	for k, v := range expect.Fields {
		switch p := v.(type) {
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
