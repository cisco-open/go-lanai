package codegen

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator"
	"embed"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
)

const (
	CommandName = "codegen"
)

var (
	Cmd = &cobra.Command{
		Use:                CommandName,
		Short:              "Given openapi contract, generate controllers/structs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}

	Args = Arguments{}
)

type Arguments struct {
	Contract          string `flag:"contract,c" desc:"openapi contract"`
	ProjectName       string `flag:"project,p" desc:"project name"`
	TemplateDirectory string `flag:"templateDir,t" desc:"Directory where templates are stored, will use built-in templates if not set"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

//go:embed all:template/src
var DefaultFS embed.FS

func Run(cmd *cobra.Command, _ []string) error {
	openAPIData, err := openapi3.NewLoader().LoadFromFile(Args.Contract)
	if err != nil {
		return fmt.Errorf("error parsing OpenAPI file: %v", err)
	}

	projectName := Args.ProjectName

	// Populate the data the templates will use
	data := map[string]interface{}{
		generator.OpenAPIData: openAPIData,
		generator.ProjectName: projectName,
	}

	FSToUse := determineFSToUse()

	template, err := generator.LoadTemplates(FSToUse)
	if err != nil {
		return err
	}

	if err = generator.GenerateFiles(
		FSToUse,
		generator.WithData(data),
		generator.WithFS(FSToUse),
		generator.WithTemplate(template)); err != nil {
		return err
	}

	fmt.Printf("Code generated to %v", cmdutils.GlobalArgs.OutputDir)
	return nil
}

func determineFSToUse() fs.FS {
	var FSToUse fs.FS
	FSToUse = DefaultFS
	if Args.TemplateDirectory == "" {
		fmt.Println("Using default template set")
	} else {
		FSToUse = os.DirFS(Args.TemplateDirectory)
	}
	return FSToUse
}
