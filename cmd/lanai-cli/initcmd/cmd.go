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

package initcmd

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"embed"
	"github.com/spf13/cobra"
)

const (
	InitRootName = "init"
	InitLibsName = "libs"
)

var (
	logger = log.New("Build.Init")
	Cmd    = &cobra.Command{
		Use:                InitRootName,
		Short:              "Initialize service, generating additional Makefile rules, Dockerfile, Docker launch script etc.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               Run,
	}
	Args = Arguments{
		Metadata: "Module.yml",
		Force:    false,
	}
	Module = ModuleMetadata{
		CliModPath: cmdutils.ModulePath,
	}
)

type Arguments struct {
	Metadata string `flag:"module-metadata,m" desc:"metadata yaml for the module"`
	Force    bool   `flag:"force,f" desc:"force overwrite files when they already exists"`
	Upgrade  bool   `flag:"upgrade" desc:"force update Makefile. Normally used together with --force"`
}

//go:embed Makefile-Build.tmpl Dockerfile.tmpl Makefile-CICD.tmpl Makefile-Libs.tmpl Makefile.tmpl dockerlaunch.tmpl
var TmplFS embed.FS

func init() {
	Cmd.AddCommand(LibInitCmd)
	cmdutils.PersistentFlags(Cmd, &Args)
}

func Run(cmd *cobra.Command, _ []string) error {
	if e := cmdutils.LoadYamlConfig(&Module, Args.Metadata); e != nil {
		return e
	}

	if e := validateModuleMetadata(cmd.Context()); e != nil {
		return e
	}

	if e := generateServiceBuildMakefile(cmd.Context()); e != nil {
		return e
	}

	if e := generateDockerfile(cmd.Context()); e != nil {
		return e
	}

	if e := generateDockerLaunchScript(cmd.Context()); e != nil {
		return e
	}

	if e := generateServiceCICDMakefile(cmd.Context()); e != nil {
		return e
	}

	if e := installBinaries(cmd.Context()); e != nil {
		return e
	}

	if !Args.Upgrade {
		return nil
	}

	if e := forceUpdateServiceMakefile(cmd.Context()); e != nil {
		return e
	}

	return nil
}
