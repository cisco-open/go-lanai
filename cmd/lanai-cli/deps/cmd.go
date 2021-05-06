package deps

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
	"strings"
)

var logger = log.New("Build.Deps")

var (
	Cmd = &cobra.Command{
		Use:                "deps",
		Short:              "dependency/go.mod related tasks",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	Args = GeneralArguments{
		Modules: []string{},
	}
)


type GeneralArguments struct {
	Modules    []string `flag:"modules,m" desc:"Comma delimited list of <module_path>[@branch]"`
	GitMarkTag string   `flag:"mark" desc:"Local git tag to mark the changes. Scripts can go back to the result of this step by using the same tag"`
}

func init() {
	Cmd.AddCommand(UpdateDepCmd)
	Cmd.AddCommand(DropReplaceCmd)
	cmdutils.PersistentFlags(Cmd, &Args)
}

func markChangesIfRequired(ctx context.Context, msg string, matchers...cmdutils.GitFileMatcher) (string, error) {
	tag := strings.TrimSpace(Args.GitMarkTag)
	if tag == "" {
		logger.WithContext(ctx).Debugf("Marking changes is not requested")
		return "", nil
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return "", e
	}
	gitutils = gitutils.WithContext(ctx)

	return tag, gitutils.MarkWorktree(tag, msg, true, matchers...)
}