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
	SourceTag    string   `flag:"src-tag,s,required" desc:"the source tag name the re-tagging is based off"`
	// TODO annotated tag
}

func init() {
	cmdutils.PersistentFlags(MarkReTagCmd, &MarkReTagArgs)
}

func RunReTagMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	src := strings.TrimSpace(MarkReTagArgs.SourceTag)
	if tag == "" || src == "" {
		return fmt.Errorf("tag and src-tag are required flags and cannot be empty")
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	if e := gitutils.TagMarkedCommit(src, tag, nil); e != nil {
		return e
	}
	logger.WithContext(cmd.Context()).Infof(`Marked tag [%s] is re-tagged as [%s]`, src, tag)
	return nil
}