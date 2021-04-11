package webjars

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/build/cmdutils"
	"encoding/json"
	"github.com/spf13/cobra"
	"os"
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
	GroupId   string  `flag:"group,g,required" desc:"Webjar's Group ID"`
	ArtifactId   string  `flag:"artifact,a,required" desc:"Webjar's Artifact ID"`
	Version   string  `flag:"version,v,required" desc:"Webjar's Version"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(cmd *cobra.Command, args []string) error {
	_ = json.NewEncoder(os.Stdout).Encode(Args)
	return nil
}
