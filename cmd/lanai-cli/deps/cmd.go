package deps

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
)

var logger = log.New("deps")

var (
	Cmd = &cobra.Command{
		Use:                "deps",
		Short:              "dependency/go.mod related tasks",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
)

func init() {
	Cmd.AddCommand(UpdateDepCmd)
	Cmd.AddCommand(DropReplaceCmd)
}