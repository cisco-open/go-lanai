package webjars

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"embed"
	"github.com/spf13/cobra"
)

const (
	defaultWebjarContentPath = "META-INF/resources/webjars"
)

var (
	Cmd = &cobra.Command{
		Use:                "webjars",
		Short:              "DownloadHandlerFunc Webjars and extract",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{
		Resources: []string{},
	}
)

type Arguments struct {
	GroupId    string   `flag:"group,g,required" desc:"Webjar's Group ID"`
	ArtifactId string   `flag:"artifact,a,required" desc:"Webjar's Artifact ID"`
	Version    string   `flag:"version,v,required" desc:"Webjar's Version"`
	Resources  []string `flag:"resources,r" desc:"Comma delimited list of additional resources from unpacked webjar. META-INF/resources/webjars is implicit"`
}

//go:embed pom.xml.tmpl
var TmplFS embed.FS

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(cmd *cobra.Command, _ []string) error {
	if e := generatePom(cmd.Context()); e != nil {
		return e
	}

	if e := executeMaven(cmd.Context()); e != nil {
		return e
	}
	return nil
}
