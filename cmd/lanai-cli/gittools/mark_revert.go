package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	MarkRevertCmd = &cobra.Command{
		Use:                "revert",
		Short:              "revert to the commit marked by given tag, without changing current branch",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunRevertMark,
	}
	MarkRevertArgs = MarkRevertArguments{}
)

type MarkRevertArguments struct {
	DiscardLocal    bool   `flag:"force-discard-changes,f" desc:"when set, local changes are discarded"`
}

func init() {
	cmdutils.PersistentFlags(MarkRevertCmd, &MarkRevertArgs)
}

func RunRevertMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	if tag == "" {
		return fmt.Errorf("tag is required flag and cannot be empty")
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	if e := gitutils.ResetToMarkedCommit(tag, MarkRevertArgs.DiscardLocal); e != nil {
		return fmt.Errorf("unable to revert to marked commit with tag %s: %v", tag, e)
	}

	logger.WithContext(cmd.Context()).Infof(`Reverted to commit marked as tag [%s]`, tag)
	return nil
}