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