package codegen

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
	"io/fs"
	"os"
	"path/filepath"
)

const (
	CommandName = "codegen"
)

var (
	logger = log.New("Codegen")
	Cmd    = &cobra.Command{
		Use:                CommandName,
		Short:              "Given openapi contract, generate controllers/structs",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args          = Arguments{}
	Configuration = Config{}
)

type Arguments struct {
	Config string `flag:"config,c" desc:"Configuration file, if not defined will default to codegen.yml"`
}

type Regeneration struct {
	Default string            `yaml:"default"`
	Rules   map[string]string `yaml:"rules"`
}
type Config struct {
	Contract           string            `yaml:"contract"`
	ProjectName        string            `yaml:"projectName"`
	TemplateDirectory  string            `yaml:"templateDirectory"`
	RepositoryRootPath string            `yaml:"repositoryRootPath"`
	Regeneration       Regeneration      `yaml:"regeneration"`
	Regexes            map[string]string `yaml:"regexes"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

//go:embed all:template/src
var DefaultFS embed.FS

func Run(cmd *cobra.Command, _ []string) error {
	ctxLog := logger.WithContext(cmd.Context())

	currDir, err := os.Getwd()
	if err != nil {
		return err
	}
	ctxLog.Debugf("Current dir: %s", currDir)

	configFilePath := Args.Config
	if configFilePath == "" {
		configFilePath = "codegen.yml"
	}
	ctxLog.Debugf("Config file path: %s", configFilePath)

	configDir := filepath.Dir(configFilePath)
	ctxLog.Debugf("Config directory: %s", configDir)

	if _, err := os.Stat(configFilePath); err == nil {
		err := processConfigurationFile(configFilePath)
		if err != nil {
			return err
		}
	}

	contractFilePath := Configuration.Contract
	ctxLog.Debugf("Configured contract file path: %s", contractFilePath)

	if !filepath.IsAbs(contractFilePath) {
		// contractFilePath is converted to be relative to current directory
		contractFilePath = filepath.Join(configDir, contractFilePath)
	}
	ctxLog.Debugf("Contract file path: %s", contractFilePath)

	openAPIData, err := openapi3.NewLoader().LoadFromFile(contractFilePath)
	if err != nil {
		return fmt.Errorf("error parsing OpenAPI file: %v", err)
	}

	projectName := Configuration.ProjectName
	repository := Configuration.RepositoryRootPath
	// Populate the data the templates will use
	data := map[string]interface{}{
		generator.OpenAPIData: openAPIData,
		generator.ProjectName: projectName,
		generator.Repository:  repository,
	}

	FSToUse := determineFSToUse(cmd.Context(), configDir)

	loaderOpts := generator.LoaderOptions{
		InitialRegexes: Configuration.Regexes,
	}
	template, err := generator.LoadTemplates(FSToUse, loaderOpts)
	if err != nil {
		return err
	}
	if err = generator.GenerateFiles(
		FSToUse,
		generator.WithData(data),
		generator.WithFS(FSToUse),
		generator.WithTemplate(template),
		generator.WithRegenerationRule(Configuration.Regeneration.Default),
		generator.WithRules(Configuration.Regeneration.Rules)); err != nil {
		return err
	}

	fmt.Printf("Code generated to %v\n", cmdutils.GlobalArgs.OutputDir)
	//	Run go mod tidy
	err = cmdutils.GoModTidy(cmd.Context(), []cmdutils.ShCmdOptions{cmdutils.ShellUseOutputDir()})
	if err != nil {
		return fmt.Errorf("could not tidy go code: %v", err)
	}
	return nil
}

func processConfigurationFile(configFilePath string) error {
	configFile, err := os.ReadFile(configFilePath)
	if err != nil {
		fmt.Printf("error parsing config file: %v\n", err)
	}
	if configFile != nil {
		config := Config{}
		err = yaml.Unmarshal(configFile, &config)
		if err != nil {
			return fmt.Errorf("error unmarshalling yaml file: %v", err)
		}
		Configuration.ProjectName = config.ProjectName
		Configuration.Contract = config.Contract
		Configuration.RepositoryRootPath = config.RepositoryRootPath
		Configuration.TemplateDirectory = config.TemplateDirectory
		Configuration.Regeneration = config.Regeneration
		Configuration.Regexes = config.Regexes
	}
	return nil
}

func determineFSToUse(ctx context.Context, configDir string) fs.FS {
	ctxLog := logger.WithContext(ctx)
	var FSToUse fs.FS
	FSToUse = DefaultFS
	templateDirectory := Configuration.TemplateDirectory
	ctxLog.Debugf("Configured template directory: %s", templateDirectory)
	if templateDirectory == "" {
		ctxLog.Debugf("Using default template set as configured template directory is empty.")
	} else {
		if !filepath.IsAbs(templateDirectory) {
			// templateDirectory is converted to be relative to current directory
			templateDirectory = filepath.Join(configDir, templateDirectory)
		}
		ctxLog.Debugf("Template directory: %s", templateDirectory)
		FSToUse = os.DirFS(templateDirectory)
	}
	return FSToUse
}
