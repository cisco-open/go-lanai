package codegen

import (
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

var logger = log.New("Codegen")

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
	Args          = Arguments{}
	Configuration = ConfigV2{}
)

type Arguments struct {
	Config string `flag:"config,c" desc:"Configuration file, if not defined will default to codegen.yml"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

//go:embed all:template/src
var DefaultFS embed.FS

func Run(cmd *cobra.Command, _ []string) error {
	configFilePath := Args.Config
	if configFilePath == "" {
		configFilePath = "codegen.yml"
	}
	if _, err := os.Stat(configFilePath); err == nil {
		err := processConfigurationFile(configFilePath)
		if err != nil {
			return err
		}
	}

	FSToUse := determineTemplateFSToUse()
	data, e := prepareContextData()
	if e != nil {
		return e
	}

	loaderOpts := generator.LoaderOptions{
		InitialRegexes: Configuration.Components.Contract.Naming.RegExps,
	}
	template, err := generator.LoadTemplates(FSToUse, loaderOpts)
	if err != nil {
		return err
	}
	if err = generator.GenerateFiles(
		generator.WithData(data),
		generator.WithOutputFS(os.DirFS(cmdutils.GlobalArgs.OutputDir)),
		generator.WithTemplateFS(FSToUse),
		generator.WithTemplate(template),
		Configuration.Regen.AsGeneratorOption()); err != nil {
		return err
	}

	logger.Infof("Code generated to %v", cmdutils.GlobalArgs.OutputDir)
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
		var versioned VersionedConfig
		err = yaml.Unmarshal(configFile, &versioned)
		if err != nil {
			return fmt.Errorf("error unmarshalling yaml file: %v", err)
		}
		if e := resolveVersionedConfig(&versioned); e != nil {
			return e
		}

		configDir := filepath.Dir(configFilePath)
		// Contract and template paths are converted to be relative to config file path
		Configuration.Components.Contract.Path = tryResolveRelativePath(Configuration.Components.Contract.Path, configDir)
		Configuration.Templates.Path = tryResolveRelativePath(Configuration.Templates.Path, configDir)
	}
	return nil
}

func resolveVersionedConfig(versioned *VersionedConfig) error {
	switch versioned.Version {
	case Version1, "":
		Configuration = *convertConfigV1(&versioned.Config)
	case Version2:
		Configuration = versioned.ConfigV2
	default:
		return fmt.Errorf(`unsupported config version "%s"`, versioned.Version)
	}
	return nil
}

func convertConfigV1(cfg *Config) *ConfigV2 {
	regenRules := make([]RegenRule, 0, len(cfg.Regeneration.Rules))
	for k, v := range cfg.Regeneration.Rules {
		regenRules = append(regenRules, RegenRule{
			Pattern: k,
			Mode:    RegenMode(v),
		})
	}
	return &ConfigV2{
		Project:    ProjectV2{
			Name:   cfg.ProjectName,
			Module: cfg.RepositoryRootPath,
		},
		Templates:  TemplatesV2{
			Path: cfg.TemplateDirectory,
		},
		Components: ComponentsV2{
			Contract: ContractV2{
				Path:   cfg.Contract,
				Naming: ContractNamingV2{
					RegExps: cfg.Regexes,
				},
			},
		},
		Regen:      RegenerationV2{
			Default: RegenMode(cfg.Regeneration.Default),
			Rules:   regenRules,
		},
	}
}

func tryResolveRelativePath(path, refDir string) string {
	if len(path) !=0 && !filepath.IsAbs(path) {
		// path is converted to be relative to refDir
		return filepath.Clean(filepath.Join(refDir, path))
	}
	return path
}

func determineTemplateFSToUse() fs.FS {
	var FSToUse fs.FS
	FSToUse = DefaultFS
	if len(Configuration.Templates.Path) == 0 {
		logger.Infof("Using default template set")
	} else {
		FSToUse = os.DirFS(Configuration.Templates.Path)
	}
	return FSToUse
}

func prepareContextData() (map[string]interface{}, error) {
	data := map[string]interface{}{
		generator.CKProjectName: Configuration.Project.Name,
		generator.CKRepository:  Configuration.Project.Module,
		generator.CKProject: generator.Project{
			Name:        Configuration.Project.Name,
			Module:      Configuration.Project.Module,
			Description: Configuration.Project.Description,
			Port:        Configuration.Project.Port,
			ContextPath: Configuration.Project.ContextPath,
		},
	}

	openAPIData, err := openapi3.NewLoader().LoadFromFile(Configuration.Components.Contract.Path)
	if err != nil {
		return nil, fmt.Errorf("error parsing OpenAPI file: %v", err)
	}
	data[generator.CKOpenAPIData] = openAPIData

	return data, nil
}