package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	MarkCreateCmd = &cobra.Command{
		Use:                "create",
		Short:              "Mark current commit with given tag name. optionally commit local changes before mark",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunCreateMark,
	}
	MarkCreateArgs = MarkCreateArguments{
		CommitMsg: "changes for git tagging",
	}
)

type MarkCreateArguments struct {
	Commit    bool   `flag:"include-local-changes" desc:"when set, local changes in worktree are commited before creating the mark"`
	CommitMsg string `flag:"commit-message,m" desc:"message used to commit local changes; ignored when 'include-local-changes' is not set. "`
}

func init() {
	cmdutils.PersistentFlags(MarkCreateCmd, &MarkCreateArgs)
}

func RunCreateMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	if tag == "" {
		return fmt.Errorf("tag is required flag and cannot be empty")
	}


	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	// commit and mark
	if MarkCreateArgs.Commit {
		logger.WithContext(cmd.Context()).Infof(`All changes are committed and as tag [%s]`, tag)
		return gitutils.MarkWorktree(tag, MarkCreateArgs.CommitMsg)
	}

	// mark head
	hash, e := gitutils.HeadCommitHash()
	if e != nil {
		return fmt.Errorf("unable to get HEAD commit hash: %v", e)
	}

	if e := gitutils.MarkCommit(tag, hash); e != nil {
		return e
	}
	logger.WithContext(cmd.Context()).Infof(`Current HEAD is marked as tag [%s]`, tag)
	return nil
}
