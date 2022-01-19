package apidocs

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/spf13/cobra"
)

/**********************
	Command
 **********************/

var (
	ResolveCmd = &cobra.Command{
		Use:                "resolve <space_delimited_source_files>",
		Short:              "Generate OAS3 document from contract definitions and external references",
		Example:            `lanai-cli apidocs resolve -o configs/api-docs.yaml -T $GITHUB_TOKEN -R https://api.swaggerhub.com/domains/Cisco-Systems46/msx-common-domain/8=github://cto-github.cisco.com/raw/NFV-BU/msx-platform-specs/v1.0.8/common-domain-8.yaml contracts/service-8.yaml contracts/service-1.json`,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args:               ValidatePositionalArgs,
		RunE:               RunMerge,
	}
	ResolveArgs = ResolveArguments{
		Output:     "api-docs-merged.yaml",
		KeepExtRef: false,
	}
)

type ResolveArguments struct {
	Output            string   `flag:"output,o" desc:"Path of output file, relative to working directory"`
	KeepExtRef        bool     `flag:"keep-external-ref,k" desc:"Keep external $ref as-is (skip resolving external $ref)"`
	GitHubPATs        []string `flag:"github-token,T" desc:"GitHub's Personal Access Token(PAT). \nFormat: \"<token>[@<hostname>]\". \nWhen <hostname> is not specified, the token is used as default."`
	ReplaceExtSources []string `flag:"replace-external-source,R" desc:"Replace external reference sources with an alternative location. \nFormat: \"<original_url>=<replaced_loc>\". \nSupported <replaced_loc> are 'http://', 'https://', 'github://' and local file. \ne.g. https://api.swaggerhub.com/domains/Cisco-Systems46/msx-common-domain/8=github://cto-github.cisco.com/raw/NFV-BU/msx-platform-specs/v1.0.8/common-domain-8.yaml"`
}

func init() {
	cmdutils.PersistentFlags(ResolveCmd, &ResolveArgs)
}

func ValidatePositionalArgs(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing source files, at least 1 source file")
	}
	return nil
}

func RunMerge(cmd *cobra.Command, args []string) error {

	logger.Infof("Loading API documents...")
	docs, e := loadApiDocs(cmd.Context(), args)
	if e != nil {
		return e
	}

	if !ResolveArgs.KeepExtRef {
		logger.Infof("Resolving external $ref...")
		docs, e = tryResolveExtRefs(cmd.Context(), docs)
		if e != nil {
			return e
		}
	}

	logger.Infof("Merging...")
	merged, e := merge(cmd.Context(), docs)
	if e != nil {
		return e
	}

	logger.Infof("Writing to [%s]...", ResolveArgs.Output)
	return writeMergedToFile(cmd.Context(), merged)
}

/************************
	Config
 ************************/

//type ResolveConfig struct {
//	GitHubTokens []string
//	ReplaceP
//}
