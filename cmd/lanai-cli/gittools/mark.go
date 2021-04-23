package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/spf13/cobra"
)

var (
	MarkCmd = &cobra.Command{
		Use:                "mark",
		Short:              "tools to create/revert/retag commit marks using a lightweight tag",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	MarkArgs = MarkArguments{}
)

type MarkArguments struct {
	MarkTag string   `flag:"tag,t,required" desc:"Local git tag to mark the worktree. Scripts can go back to the result of this step by using the same tag"`
}

func init() {
	MarkCmd.AddCommand(MarkCreateCmd)
	MarkCmd.AddCommand(MarkRevertCmd)
	MarkCmd.AddCommand(MarkReTagCmd)
	cmdutils.PersistentFlags(MarkCmd, &MarkArgs)
}

