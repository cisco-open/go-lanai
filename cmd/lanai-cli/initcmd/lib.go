package initcmd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/spf13/cobra"
)

var (
	LibInitCmd    = &cobra.Command{
		Use:                InitLibsName,
		Short:              "Initialize library module, generating additional Makefile rules, etc.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunLibsInit,
	}
)

func RunLibsInit(cmd *cobra.Command, _ []string) error {
	if e := cmdutils.LoadYamlConfig(&Module, Args.Metadata); e != nil {
		return e
	}

	if e := validateModuleMetadata(cmd.Context()); e != nil {
		return e
	}

	if e := generateLibsCICDMakefile(cmd.Context()); e != nil {
		return e
	}

	return nil
}
