package codegen

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"fmt"
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
)

type Arguments struct {
	Config string `flag:"config,c" desc:"Configuration file, if not defined will default to codegen.yml"`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

const DefaultTemplateRoot = "template/src"

//go:embed all:template/src
var DefaultTemplateFS embed.FS

func Run(cmd *cobra.Command, _ []string) error {
	// process arguments
	configFilePath := Args.Config
	if configFilePath == "" {
		configFilePath = "codegen.yml"
	}

	// do generation
	if e := GenerateWithConfigPath(cmd.Context(), configFilePath); e != nil {
		return e
	}

	//	run go mod tidy

	if e := cmdutils.GoModTidy(cmd.Context(), []cmdutils.ShCmdOptions{cmdutils.ShellUseOutputDir()}); e != nil {
		return fmt.Errorf("'go mod tidy' failed: %v", e)
	}
	return nil
}

func GenerateWithConfigPath(ctx context.Context, configPath string) error {
	if _, e := os.Stat(configPath); e != nil {
		return e
	}
	cfg, e := processConfigurationFile(configPath)
	if e != nil {
		return e
	}
	return GenerateWithConfig(ctx, cfg)
}

func GenerateWithConfig(ctx context.Context, cfg *ConfigV2) error {
	tmplFS, e := determineTemplateFSToUse(cfg)
	if e != nil {
		return e
	}
	loaderOpts := generator.LoaderOptions{
		InitialRegexes: cfg.Components.Contract.Naming.RegExps,
	}
	template, err := generator.LoadTemplates(tmplFS, loaderOpts)
	if err != nil {
		return err
	}

	// Do generate
	opts := append(cfg.ToOptions(),
		generator.WithTemplateFS(tmplFS),
		generator.WithTemplate(template),
	)
	if err = generator.GenerateFiles(opts...); err != nil {
		return err
	}
	logger.Infof("Code generated to %v", cmdutils.GlobalArgs.OutputDir)
	return nil
}

func processConfigurationFile(configFilePath string) (*ConfigV2, error) {
	configFile, e := os.ReadFile(configFilePath)
	if e != nil {
		return nil, fmt.Errorf(`error reading config file "%s": %v`, configFilePath, e)
	}
	var versioned VersionedConfig
	e = yaml.Unmarshal(configFile, &versioned)
	if e != nil {
		return nil, fmt.Errorf(`error unmarshalling yaml file "%s": %v`, configFilePath, e)
	}
	cfg, e := resolveVersionedConfig(&versioned)
	if e != nil {
		return nil, e
	}

	configDir := filepath.Dir(configFilePath)
	// Contract and template paths are converted to be relative to config file path
	cfg.Components.Contract.Path = tryResolveRelativePath(cfg.Components.Contract.Path, configDir)
	cfg.Templates.Path = tryResolveRelativePath(cfg.Templates.Path, configDir)
	return cfg, nil
}

func resolveVersionedConfig(versioned *VersionedConfig) (*ConfigV2, error) {
	switch versioned.Version {
	case Version1, "":
		return versioned.Config.ToV2(), nil
	case Version2:
		return &versioned.ConfigV2, nil
	default:
		return nil, fmt.Errorf(`unsupported config version "%s"`, versioned.Version)
	}
}

func tryResolveRelativePath(path, refDir string) string {
	if len(path) != 0 && !filepath.IsAbs(path) {
		// path is converted to be relative to refDir
		return filepath.Clean(filepath.Join(refDir, path))
	}
	return path
}

func determineTemplateFSToUse(cfg *ConfigV2) (fs.FS, error) {
	if len(cfg.Templates.Path) == 0 {
		logger.Infof("Using default template set")
		return fs.Sub(DefaultTemplateFS, DefaultTemplateRoot)
	} else {
		return os.DirFS(cfg.Templates.Path), nil
	}
}
