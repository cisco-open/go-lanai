package dev

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
)

var logger = log.New("Build.Dev")

var (
	Cmd = &cobra.Command{
		Use:                "dev",
		Short:              "helper utilities for developers",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
)




func init() {
	//Cmd.AddCommand(UpdateDepCmd)
	//Cmd.AddCommand(DropReplaceCmd)
	Cmd.AddCommand(AddReplaceCmd)
	cmdutils.PersistentFlags(Cmd, &AddReplaceArgs)
}
