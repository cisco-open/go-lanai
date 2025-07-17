package cmdtest

import (
	"context"
	"fmt"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/cisco-open/go-lanai/test"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"testing"
)

const (
	TestTmpDir    = ".tmp"
	TestOutputDir = "output"
)

var (
	testRootCmd = &cobra.Command{
		Use:                "lanai-cli-test",
		Short:              "lanai-cli for test",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PersistentPreRunE: cmdutils.MergeRunE(
			cmdutils.EnsureGlobalDirectories(),
			cmdutils.PrintEnvironment(),
		),
	}
)

func init() {
	cmdutils.PersistentFlags(testRootCmd, &cmdutils.GlobalArgs)
}

func SetupResetPackageVars() test.SetupFunc {
	return func(ctx context.Context, t *testing.T) (context.Context, error) {
		cmdutils.ResetGoCmd()
		return ctx, nil
	}
}

// DryRunCobraCommand Run given cobra.Command in given work dir relative to "testdata"
func DryRunCobraCommand(ctx context.Context, wd string, cmd *cobra.Command, handler TestDryRunHandler, args ...string) error {
	// prepare global args
	originalArgs := cmdutils.GlobalArgs
	defer func() { cmdutils.GlobalArgs = originalArgs }()

	cmdutils.GlobalArgs = cmdutils.Global{
		WorkingDir: PathRelativeToTestdata(wd),
		TmpDir:     PathRelativeToTestdata(wd, TestTmpDir),
		OutputDir:  PathRelativeToTestdata(wd, TestOutputDir),
		Verbose:    true,
		DryRun:     true,
	}

	// setup dry run handler
	if handler != nil {
		cmdutils.GlobalArgs.DryRunFunc = handler.Handle
	} else {
		cmdutils.GlobalArgs.DryRunFunc = originalArgs.DryRunFunc
	}

	// clean up FS
	if e := os.RemoveAll(cmdutils.GlobalArgs.OutputDir); e != nil {
		return fmt.Errorf("unable to start command with clean FS: %w", e)
	}

	// run command
	if len(args) == 0 {
		args = []string{}
	}
	cmdCopy, _ := CopyCommandChain(cmd, args...)
	return cmdCopy.ExecuteContext(ctx)
}

// CopyCommandChain do following things:
// - Make copy of given command and its parents/ancestors.
// - Fix arguments of its ancestors.
// - Attach the given command chain to a copy of predefined test root command (to mimic how main() function works)
// This function returns copied command and a copy of its root command. Two values may be same if they are the root
func CopyCommandChain(cmd *cobra.Command, args...string) (cmdCpy, rootCpy *cobra.Command) {
	cmdCpy = copyCmd(cmd)
	var cpy, prev *cobra.Command
	for cpy, prev = cmdCpy, nil; cpy != nil; cpy = copyCmd(cpy.Parent()) {
		if prev != nil {
			args = append([]string{prev.Name()}, args...)
			cpy.ResetCommands()
			cpy.AddCommand(prev)
		}
		cpy.SetArgs(args)
		prev = cpy
	}
	rootCpy = copyCmd(testRootCmd)
	if prev != nil {
		args = append([]string{prev.Name()}, args...)
		rootCpy.ResetCommands()
		rootCpy.AddCommand(prev)
	}
	rootCpy.SetArgs(args)
	return
}

func copyCmd(cmd *cobra.Command) *cobra.Command {
	if cmd == nil {
		return nil
	}
	vCopy := *cmd
	return &vCopy
}

func PathRelativeToTestdata(pathComponents ...string) string {
	if base, e := os.Getwd(); e != nil {
		pathComponents = append([]string{"testdata"}, pathComponents...)
	} else {
		pathComponents = append([]string{base, "testdata"}, pathComponents...)
	}
	return filepath.Join(pathComponents...)
}
