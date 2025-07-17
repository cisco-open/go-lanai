package initcmd

import (
	"context"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdtest"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

var (
	ExpectedLibOutputs = []string{
		"Makefile-Generated",
	}
)

/*************************
	Tests
 *************************/

func TestInitLibsCommand(t *testing.T) {
	test.RunTest(context.Background(), t,
		cmdtest.WithDryRun(InitDryRunHandler),
		test.SubTestSetup(SetupResetInitPackageVariables()),
		test.GomegaSubTest(SubTestInitLibsWithoutError(), "WithoutError"),
		test.GomegaSubTest(SubTestInitLibsWithConflictBinaries(), "WithConflictBinaries"),
		test.GomegaSubTest(SubTestInitLibsWithNonExistExecutable(), "WithNonExistExecutable"),
		test.GomegaSubTest(SubTestInitLibsWithNonExistGenerates(), "WithNonExistGenerates"),
		test.GomegaSubTest(SubTestInitLibsWithNonExistSources(), "WithNonExistSources"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestInitLibsWithoutError() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, LibInitCmd, nil, "-m", "Module-valid.yml",)
		g.Expect(e).To(Succeed(), "command should succeed")
		ExpectOutputFiles(g, TestdataSimpleGoMod, false, ExpectedLibOutputs...)
	}
}

func SubTestInitLibsWithConflictBinaries() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, LibInitCmd, nil, "-m", "Module-invalid-binaries.yml",)
		ExpectFailure(g, e)
	}
}

func SubTestInitLibsWithNonExistExecutable() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-execs.yml")
		ExpectFailure(g, e)
	}
}

func SubTestInitLibsWithNonExistGenerates() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-generates.yml")
		ExpectFailure(g, e)
	}
}

func SubTestInitLibsWithNonExistSources() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		e := cmdtest.DryRunCobraCommand(ctx, TestdataSimpleGoMod, Cmd, nil, "-m", "Module-invalid-sources.yml")
		// NOTE: "sources" can be non-exists.
		// Currently, we don't enforce that, coz when "lanai-cli init" is run, extra sources might not exist yet
		g.Expect(e).To(Succeed(), "command should succeed")
		ExpectOutputFiles(g, TestdataSimpleGoMod, false, ExpectedLibOutputs...)
	}
}

