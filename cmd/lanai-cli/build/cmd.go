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

package build

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"time"
)

var (
	Cmd = &cobra.Command{
		Use:                "build [--version|-v version] [--ldflags additional_ldflags] -- [other 'go build' arguments]",
		Short:              "utilities to help build project with proper ldflags and mods",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE: RunBuild,
	}
	Args = GeneralArguments{
		Version: "Unknown",
	}
)

type GeneralArguments struct {
	Version string   `flag:"version,v" desc:"Version value to be included as 'version' in build info"`
	Modules []string `flag:"deps,d" desc:"Comma delimited list of <module_path> to be included as 'dependencies'' in build info. Note: if <module_path> is followed by [@branch], the branch value is ignored"`
	LdFlags string   `flag:"ldflags" desc:"Additional ldflags passed to \"go build\""`
}

func init() {
	cmdutils.PersistentFlags(Cmd, &Args)
}

func RunBuild(cmd *cobra.Command, args []string) error {
	// calculate build info ldflags
	ldflags := strings.Join([]string {
		ldFlagsForBuildInfo(cmd.Context()),
		Args.LdFlags,
	}, " ")

	// go build
	shcmd := fmt.Sprintf(`go build -ldflags="%s" %s`, ldflags, strings.Join(args, " "))
	_, e := cmdutils.RunShellCommands(cmd.Context(),
		cmdutils.ShellShowCmd(true),
		cmdutils.ShellUseWorkingDir(),
		cmdutils.ShellCmd(shcmd),
		cmdutils.ShellStdOut(os.Stdout),
		)
	return e
}

func ldFlagsForBuildInfo(ctx context.Context) string {

	flags := []string {
		ldFlagBootstrapVariable("BuildVersion", Args.Version),
		ldFlagBootstrapVariable("BuildTime", time.Now().Format(utils.ISO8601Seconds)),
		ldFlagBootstrapVariable("BuildHash", gitHeadHash(ctx)),
		ldFlagBootstrapVariable("BuildDeps", moduleDeps(ctx)),
	}

	return strings.Join(flags, " ")
}

// ldFlagBuildInfoVariable work with bootstrap.*
func ldFlagBootstrapVariable(varName string, v string ) string {
	if v == "" {
		return ""
	}
	return fmt.Sprintf("-X '%s/pkg/bootstrap.%s=%s'", cmdutils.ModulePath, varName, v)
}

func gitHeadHash(ctx context.Context) string {
	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return ""
	}

	hash, e := gitutils.WithContext(ctx).HeadCommitHash()
	if e != nil {
		return ""
	}
	return hash.String()
}

func moduleDeps(ctx context.Context) string {
	modPaths := utils.NewStringSet(cmdutils.ModulePath)
	for _, m := range Args.Modules {
		path := strings.TrimSpace(strings.SplitN(m, "@", 2)[0])
		if path != "" {
			modPaths.Add(path)
		}
	}

	mods, e := cmdutils.FindModule(ctx, nil, modPaths.Values()...)
	if e != nil {
		return ""
	}

	var modules []string
	for _, m := range mods {
		ver := m.Version
		switch {
		case m.Replace != nil && m.Replace.Version != "":
			ver = m.Replace.Version
		case m.Replace != nil:
			ver = "0.0.0-SNAPSHOT"
		case ver == "":
			ver = "0.0.0-UNKNOWN"
		}
		modules = append(modules, fmt.Sprintf("%s@%s", m.Path, ver))
	}
	return strings.Join(modules, ",")
}