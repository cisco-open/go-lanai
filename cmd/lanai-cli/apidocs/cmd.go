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
	Cmd.AddCommand(MergeCmd)
}

func Run(cmd *cobra.Command, _ []string) error {
	//confPath := filepath.Join(Args.Source, "ccv.yml")
	//conf, e := config.LoadFrom(confPath)
	//if e != nil {
	//	logger.Errorf(`unable to load config file "%s"`, confPath)
	//	return e
	//}
	//
	//out := "./contract.yml"
	//data, e := conf.ResolveContracts()
	//if e != nil {
	//	return e
	//}
	//
	//if ext := filepath.Ext(out); !shared.IsYamlExtension(ext) {
	//	if data, e = shared.ConvertYamlTo(data, ext); e != nil {
	//		return e
	//	}
	//}
	//
	//if e := shared.WriteFile(out, data, 0777); e != nil {
	//	return e
	//}

	//if e := cmdutils.LoadYamlConfig(&Module, Args.Metadata); e != nil {
	//	return e
	//}
	//
	//if e := validateModuleMetadata(cmd.Context()); e != nil {
	//	return e
	//}
	//
	//if e := generateServiceBuildMakefile(cmd.Context()); e != nil {
	//	return e
	//}
	//
	//if e := generateDockerfile(cmd.Context()); e != nil {
	//	return e
	//}
	//
	//if e := generateDockerLaunchScript(cmd.Context()); e != nil {
	//	return e
	//}
	//
	//if e := generateServiceCICDMakefile(cmd.Context()); e != nil {
	//	return e
	//}
	//
	//if !Args.Upgrade {
	//	return nil
	//}
	//
	//if e := forceUpdateServiceMakefile(cmd.Context()); e != nil {
	//	return e
	//}

	return nil
}
