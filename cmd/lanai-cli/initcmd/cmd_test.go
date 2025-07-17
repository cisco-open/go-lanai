package initcmd

import (
	"context"
	"errors"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdtest"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/cisco-open/go-lanai/pkg/utils/matcher"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

var InitDryRunHandler = cmdtest.TestDryRunHandler{
	cmdutils.DryRunTypeShell: cmdtest.DryRunShellWithFilter(true,
		matcher.WithPrefix("go install", false),
	),
}

const (
	TestdataSimpleGoMod = "simple_gomod"
)

var (
	ExpectedServiceOutputs = []string{
		"Makefile-Build",
		"Makefile-Generated",
		"build/package/Dockerfile",
		"build/package/dockerlaunch.sh",
	}
)

/*************************
	Test Setup
 *************************/

func SetupResetInitPackageVariables() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		Module = ModuleMetadata{
			CliModPath: cmdutils.ModulePath,
		}
		return ctx, nil
	}
}

/*************************
	Tests
 *************************/

func TestInitCommand(t *testing.T) {
	test.RunTest(context.Background(), t,
		cmdtest.WithDryRun(InitDryRunHandler),
		test.SubTestSetup(SetupResetInitPackageVariables()),
		test.GomegaSubTest(SubTestInitSvcWithoutError(), "WithoutError"),
		test.GomegaSubTest(SubTestInitSvcWithConflictBinaries(), "WithConflictBinaries"),
		test.GomegaSubTest(SubTestInitSvcWithNonExistExecutable(), "WithNonExistExecutable"),
		test.GomegaSubTest(SubTestInitSvcWithNonExistGenerates(), "WithNonExistGenerates"),
		test.GomegaSubTest(SubTestInitSvcWithNonExistSources(), "WithNonExistSources"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestInitSvcWithoutError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-valid.yml")
		g.Expect(e).To(Succeed(), "command should succeed")
		ExpectOutputFiles(g, TestdataSimpleGoMod, false, ExpectedServiceOutputs...)
	}
}

func SubTestInitSvcWithConflictBinaries() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-binaries.yml")
		ExpectFailure(g, e)
	}
}

func SubTestInitSvcWithNonExistExecutable() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-execs.yml")
		ExpectFailure(g, e)
	}
}

func SubTestInitSvcWithNonExistGenerates() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-generates.yml")
		ExpectFailure(g, e)
	}
}

func SubTestInitSvcWithNonExistSources() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-sources.yml")
		// NOTE: "sources" can be non-exists.
		// Currently, we don't enforce that, coz when "lanai-cli init" is run, extra sources might not exist yet
		g.Expect(e).To(Succeed(), "command should succeed")
		ExpectOutputFiles(g, TestdataSimpleGoMod, false, ExpectedServiceOutputs...)
	}
}

/*************************
	Helpers
 *************************/

func ExpectFailure(g *gomega.WithT, err error) {
	g.Expect(err).To(HaveOccurred(), "command should fail")
	pe := &fs.PathError{}
	g.Expect(errors.As(err, &pe)).To(BeFalse(), "failure should not caused by not finding module file")
}

// ExpectOutputFiles we don't check content here
func ExpectOutputFiles(g *gomega.WithT, wd string, allowEmpty bool, expectedFiles ...string) {
	base := cmdtest.PathRelativeToTestdata(wd, cmdtest.TestOutputDir)
	for _, filename := range expectedFiles {
		data, e := os.ReadFile(filepath.Join(base, filename))
		g.Expect(e).To(Succeed(), "file '%s' should exist in folder %s", filename, wd)
		if !allowEmpty {
			g.Expect(data).ToNot(BeEmpty(), "file '%s' should in folder %s should not be empty", filename, wd)
		}
	}
}