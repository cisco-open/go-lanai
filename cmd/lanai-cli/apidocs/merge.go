package apidocs

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"github.com/imdario/mergo"
	"github.com/spf13/cobra"
)

/**********************
	Command
 **********************/

var (
	MergeCmd = &cobra.Command{
		Use:                "merge <space_delimited_source_files>",
		Short:              "Initialize library module, generating additional Makefile rules, etc.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args:               ValidatePositionalArgs,
		RunE:               RunMerge,
	}
	MergeArgs = MergeArguments{
		Output:     "api-docs-merged.yaml",
		KeepExtRef: false,
	}
)

type MergeArguments struct {
	Output     string   `flag:"output,o" desc:"output file path, relative to working directory"`
	KeepExtRef bool     `flag:"keep-external-ref,k" desc:"keep external $ref as-is (skip resolving external $ref)"`
	GitHubPATs []string `flag:"github-token,T" desc:"GitHub's Personal Access Token(PAT) with format \"<token>[@<hostname>]\". When <hostname> is not specified, the token is used as default."`
}

func init() {
	cmdutils.PersistentFlags(MergeCmd, &MergeArgs)
}

func ValidatePositionalArgs(cmd *cobra.Command, args []string) error {
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

	// TODO resolve external $ref
	if !MergeArgs.KeepExtRef {
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

	logger.Infof("Writing to [%s]...", MergeArgs.Output)
	return writeMergedToFile(cmd.Context(), merged)
}

/**********************
	Implementation
 **********************/

func merge(_ context.Context, docs []*apidoc) (*apidoc, error) {
	merged := make(map[string]interface{})
	for _, doc := range docs {
		if e := mergo.Merge(&merged, doc.value, mergo.WithAppendSlice); e != nil {
			return nil, fmt.Errorf("failed to merge [%s]: %v", doc.source, e)
		}
	}

	return &apidoc{
		source: MergeArgs.Output,
		value:  merged,
	}, nil
}

func writeMergedToFile(ctx context.Context, doc *apidoc) error {
	doc.source = MergeArgs.Output
	absPath, e := writeApiDocLocal(ctx, doc)
	if e != nil {
		return e
	}
	logger.Infof("Merged API document saved to [%s]", absPath)
	return nil
}
