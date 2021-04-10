package webjars

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	Cmd = &cobra.Command{
		Use:                "webjars",
		Short:              "Download Webjars and extract",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{}
)

type Arguments struct {
	Value     string  `flag:"str,s,default str" desc:"string value"`
	StrPtr  *string `flag:"" desc:"stringPtr"`
	Int     int     `flag:"" desc:""`
	IntPtr  *int    `flag:"" desc:""`
	Bool    bool    `flag:"" desc:""`
	BoolPtr *bool   `flag:"" desc:""`
}

func init() {
	Cmd.PersistentFlags().StringVarP(&Args.Value, "value", "v", "default", "Test value")
}

func Run(cmd *cobra.Command, args []string) error {
	fmt.Printf("Arg: %s\n", Args.Value)
	return nil
}
