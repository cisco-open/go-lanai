// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package codegen

import (
    "context"
    "embed"
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen/generator"
    "github.com/cisco-open/go-lanai/pkg/log"
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
	if !cmdutils.GlobalArgs.Verbose {
		_ = log.UpdateLoggingConfiguration(&log.Properties{
			Levels:   map[string]log.LoggingLevel{
				"default": log.LevelInfo,
			},
		})
	}

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
	// Do generate
	opts := append(cfg.ToOptions(),
		generator.WithTemplateFS(tmplFS),
	)
	if e = generator.GenerateFiles(ctx, opts...); e != nil {
		return e
	}
	logger.Infof("Code generated to %v", cmdutils.GlobalArgs.OutputDir)
	return nil
}

func processConfigurationFile(configFilePath string) (*ConfigV2, error) {
	configFile, e := os.ReadFile(configFilePath)
	if e != nil {
		return nil, fmt.Errorf(`error reading config file "%s": %v`, configFilePath, e)
	}
	versioned := DefaultVersionedConfig
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
