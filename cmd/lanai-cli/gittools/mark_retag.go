package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	MarkReTagCmd = &cobra.Command{
		Use:                "retag",
		Short:              "mark current commit as a lightweight tag",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunReTagMark,
	}
	MarkReTagArgs = MarkReTagArguments{}
)

type MarkReTagArguments struct {
	ReTag    string   `flag:"retag,,required" desc:"tag name to re-tag"`
	// TODO annotated tag
}

func init() {
	cmdutils.PersistentFlags(MarkReTagCmd, &MarkReTagArgs)
}

func RunReTagMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	retag := strings.TrimSpace(MarkReTagArgs.ReTag)
	if tag == "" || retag == "" {
		return fmt.Errorf("tag and retag are required flags and cannot be empty")
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	if e := gitutils.TagMarkedCommit(tag, retag, nil); e != nil {
		return e
	}
	logger.WithContext(cmd.Context()).Infof(`Marked tag [%s] is re-tagged as [%s]`, tag, retag)
	return nil
}