package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/go-git/go-git/v5"
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
	SourceTag    string   `flag:"src-tag,s" desc:"the source tag name the re-tagging is based off. If not provided, current HEAD is used"`
	// TODO annotated tag
}

func init() {
	cmdutils.PersistentFlags(MarkReTagCmd, &MarkReTagArgs)
}

func RunReTagMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	src := strings.TrimSpace(MarkReTagArgs.SourceTag)
	if tag == "" {
		return fmt.Errorf("tag is required flags and cannot be empty")
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	// when opts is not nil, the result tag is annotated tag
	var opts *git.CreateTagOptions
	if src == "" {
		hash, e := gitutils.HeadCommitHash()
		if e != nil {
			return e
		}
		if e := gitutils.TagCommit(tag, hash, opts, true); e != nil {
			return e
		}
		logger.WithContext(cmd.Context()).Infof(`--src-tag is not set. Current HEAD is re-tagged as [%s]`, tag)
	} else {
		if e := gitutils.TagMarkedCommit(src, tag, opts); e != nil {
			return e
		}
		logger.WithContext(cmd.Context()).Infof(`Marked tag [%s] is re-tagged as [%s]`, src, tag)
	}

	return nil
}