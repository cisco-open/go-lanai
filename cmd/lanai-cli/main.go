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

package main

import (
	"context"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/apidocs"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/build"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/codegen"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/deps"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/dev"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/gittools"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/initcmd"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/noop"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/webjars"
	"github.com/cisco-open/go-lanai/pkg/log"
	"github.com/cisco-open/go-lanai/pkg/utils"
	"github.com/spf13/cobra"
	"os"
)

const (
	CliName = "lanai-cli"
)

var (
	BuildVersion = "unknown"
)

var (
	rootCmd = &cobra.Command{
		Use:                CliName,
		Short:              "A go-lanai CLI building tool.",
		Long:               "This is a go-lanai CLI building tool.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		PersistentPreRunE: cmdutils.MergeRunE(
			cmdutils.EnsureGlobalDirectories(),
			cmdutils.PrintEnvironment(),
		),
	}
	logTemplate = `{{pad -25 .time}} [{{lvl 5 .}}]: {{.msg}}`
	logProps    = log.Properties{
		Levels: map[string]log.LoggingLevel{
			"default": log.LevelDebug,
		},
		Loggers: map[string]*log.LoggerProperties{
			"console": {
				Type:     log.TypeConsole,
				Format:   log.FormatText,
				Template: logTemplate,
				FixedKeys: utils.CommaSeparatedSlice{
					log.LogKeyName, log.LogKeyMessage, log.LogKeyTimestamp,
					log.LogKeyCaller, log.LogKeyLevel, log.LogKeyContext,
				},
			},
		},
		Mappings: map[string]string{},
	}
)

func init() {
	if e := log.UpdateLoggingConfiguration(&logProps); e != nil {
		panic(e)
	}
}

func main() {
	rootCmd.Version = BuildVersion
	rootCmd.AddCommand(noop.Cmd)
	rootCmd.AddCommand(webjars.Cmd)
	rootCmd.AddCommand(initcmd.Cmd)
	rootCmd.AddCommand(deps.Cmd)
	rootCmd.AddCommand(gittools.Cmd)
	rootCmd.AddCommand(build.Cmd)
	rootCmd.AddCommand(apidocs.Cmd)
	rootCmd.AddCommand(codegen.Cmd)
	rootCmd.AddCommand(dev.Cmd)

	cmdutils.PersistentFlags(rootCmd, &cmdutils.GlobalArgs)
	if e := rootCmd.ExecuteContext(context.Background()); e != nil {
		os.Exit(1)
	}
}
