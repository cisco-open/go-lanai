package initcmd

import (
	"context"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/onsi/gomega"
	"testing"
)

func TestInstallBinaries(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.Setup(SetupGlobalArgs()),
		test.GomegaSubTest(SubTestWithBinaryOverride(), "WithBinaryOverride"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SetupGlobalArgs() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		cmdutils.GlobalArgs.DryRun = true
		return ctx, nil
	}
}

func SubTestWithBinaryOverride() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		Module.Binaries = []*Binary{
			{
				Package: "github.com/golangci/golangci-lint/cmd/golangci-lint",
				Version: "",
			},
			{
				Package: "github.com/golangci/golangci-lint/v2/cmd/golangci-lint",
				Version: "v2.1.4",
			},
			{
				Package: "",
				Version: "v0.0.0",
			},
		}
		e := installBinaries(ctx)
		g.Expect(e).To(gomega.Succeed())
	}
}
