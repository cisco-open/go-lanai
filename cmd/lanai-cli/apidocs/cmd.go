package apidocs

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"github.com/spf13/cobra"
)

var (
	logger = log.New("Build.APIDocs")
	Cmd    = &cobra.Command{
		Use:                "apidocs",
		Short:              "Utilities to work with OpenAPI Specs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		//RunE:               Run,
	}
)

func init() {
	Cmd.AddCommand(ResolveCmd)
}


