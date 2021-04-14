package initcmd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"github.com/spf13/cobra"
)

var (
	logger = log.New("Build")
	Cmd    = &cobra.Command{
		Use:                "init",
		Short:              "Initialize service, generating additional Makefile rules, Dockerfile, etc.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{
		Metadata: "Module.yml",
		Force:    false,
	}
	Module = ModuleMetadata{}
)

type Arguments struct {
	Metadata string `flag:"module-metadata,m" desc:"metadata yaml for the module"`
	Force    bool   `flag:"force,f" desc:"force overwrite generated file when they already exists"`
}

//go:embed Makefile-Build.tmpl Dockerfile.tmpl
var TmplFS embed.FS

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(cmd *cobra.Command, _ []string) error {
	if e := cmdutils.LoadYamlConfig(&Module, Args.Metadata); e != nil {
		return e
	}

	if e := validateModuleMetadata(cmd.Context()); e != nil {
		return e
	}

	if e := generateMakefile(cmd.Context()); e != nil {
		return e
	}

	if e := generateDockerfile(cmd.Context()); e != nil {
		return e
	}
	return nil
}
