package gittools

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
)

var logger = log.New("Build.Git")

var (
	Cmd = &cobra.Command{
		Use:                "git",
		Short:              "Git plumbing tools",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
)

func init() {
	Cmd.AddCommand(MarkCmd)
}
