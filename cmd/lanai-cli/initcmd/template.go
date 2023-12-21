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
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"path/filepath"
	"text/template"
)

type TemplateData struct {
	Package     string
	Executables map[string]Executable
	Generates   []Generate
	Resources   []Resource
}

func forceUpdateServiceMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
		Customizer: func(t *template.Template) {
			// custom delimiter, we want this file's templates shows as-is
			t.Delims("{{{", "}}}")
		},
	})
}


func generateServiceCICDMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-CICD.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Generated"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      &Module,
	})
}

func generateServiceBuildMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-Build.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Build"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}

func generateDockerfile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Dockerfile.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "build/package/Dockerfile"),
		OutputPerm: 0644,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}

func generateDockerLaunchScript(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "dockerlaunch.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "build/package/dockerlaunch.sh"),
		OutputPerm: 0755,
		Overwrite:  Args.Force,
		Model:      &Module,
	})
}

func generateLibsCICDMakefile(ctx context.Context) error {
	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "Makefile-Libs.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.OutputDir, "Makefile-Generated"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      &Module,
	})
}