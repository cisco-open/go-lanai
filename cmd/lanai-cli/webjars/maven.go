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

package webjars

import (
	"context"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"path/filepath"
)

type resource struct {
	Directory string
	Includes []string
	Excludes []string
}
type pomTmplData struct {
	*cmdutils.Global
	*Arguments
	Resources []resource
}


func generatePom(ctx context.Context) error {
	resources := []resource{
		{
			Directory: defaultWebjarContentPath,
			Excludes: []string{"vendor/**"},
		},
	}
	for _, res := range Args.Resources {
		resources = append(resources, resource{
			Directory: res,
		})
	}
	data := &pomTmplData{
		Global: &cmdutils.GlobalArgs,
		Arguments: &Args,
		Resources: resources,
	}

	return cmdutils.GenerateFileWithOption(ctx, &cmdutils.TemplateOption{
		FS:         TmplFS,
		TmplName:   "pom.xml.tmpl",
		Output:     filepath.Join(cmdutils.GlobalArgs.TmpDir, "pom.xml"),
		OutputPerm: 0644,
		Overwrite:  true,
		Model:      data,
	})
}

func executeMaven(ctx context.Context) error {
	_, e := cmdutils.RunShellCommands(ctx,
		cmdutils.ShellShowCmd(true),
		cmdutils.ShellUseTmpDir(),
		cmdutils.ShellCmd("mvn versions:resolve-ranges -q"),
		cmdutils.ShellCmd("mvn dependency:unpack-dependencies -q"),
		cmdutils.ShellCmd("mvn resources:copy-resources -q"),
	)
	return e
}

