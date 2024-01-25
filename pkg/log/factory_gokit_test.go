package log

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/go-kit/log"
	"github.com/onsi/gomega"
	"io/fs"
	"testing"
)

/*************************
	Tests
 *************************/

func TestGoKitLogger(t *testing.T) {
	factoryCreator := GoKitFactoryCreator()
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
		test.GomegaSubTest(SubTestWithCaller(factoryCreator, log.Caller(7)), "WithCaller"),
		test.GomegaSubTest(SubTestConcurrent(factoryCreator), "Concurrent"),
		test.GomegaSubTest(SubTestRefresh(factoryCreator), "Refresh"),
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

func GoKitFactoryCreator() TestFactoryCreateFunc {
	return func(g *gomega.WithT, fsys fs.FS, path string) loggerFactory {
		p := BindProperties(g, fsys, path)
		return newKitLoggerFactory(&p)
	}
}