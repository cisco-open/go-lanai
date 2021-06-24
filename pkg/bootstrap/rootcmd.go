/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package bootstrap

import (
	"fmt"
	"github.com/spf13/cobra"
	"go.uber.org/fx"
	"os"
	"regexp"
)

const (
	CliFlagActiveProfile     = "active-profiles"
	CliFlagAdditionalProfile = "additional-profiles"
	CliFlagConfigSearchPath  = "config-search-path"
)

var (
	argsPattern = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9\-._]+=.*`)
	// rootCmd represents the base command when called without any subcommands
	// Note: when running app as `./app --flag1 value1 --flag2 value2 -- any-thing...`
	// 		 the values after bare `--` are passed in as args. we could use it as CLI properties assignment
	rootCmd = &cobra.Command{
		Short:              "A go-lanai based application.",
		Long:               "This is a go-lanai based application.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if !argsPattern.MatchString(arg) {
					return fmt.Errorf(`CLI properties should be in format of "property-path=value", but got "%s"`, arg)
				}
			}
			return nil
		},
	}
	cliCtx = CliExecContext{}
)

type CliExecContext struct {
	Cmd                *cobra.Command
	ActiveProfiles     []string
	AdditionalProfiles []string
	ConfigSearchPaths  []string
	Args               []string
}

func init() {
	// config flags
	rootCmd.PersistentFlags().StringSliceVarP(&cliCtx.ActiveProfiles, CliFlagActiveProfile, "P", []string{},
		`Comma separated active profiles. Override property "application.profiles.active"`)
	rootCmd.PersistentFlags().StringSliceVarP(&cliCtx.AdditionalProfiles, CliFlagAdditionalProfile, "a", []string{},
		`Comma separated additional profiles. Set property "application.profiles.additional". Additional profiles is added to active profiles`)
	rootCmd.PersistentFlags().StringSliceVarP(&cliCtx.ConfigSearchPaths, CliFlagConfigSearchPath, "c", []string{},
		`Comma separated paths. Override property "config.file.search-path"`)
}

// AddStringFlag should be called before Execute() to register flags that are supported
func AddStringFlag(flagVar *string, name string, defaultValue string, usage string) {
	rootCmd.PersistentFlags().StringVar(flagVar, name, defaultValue, usage)
}

func AddBoolFlag(flagVar *bool, name string, defaultValue bool, usage string) {
	rootCmd.PersistentFlags().BoolVar(flagVar, name, defaultValue, usage)
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Errorf("%v", err)
		os.Exit(1)
	}
}

type CliOptions func(cmd *cobra.Command)

func NewAppCmd(appName string, priorityOptions []fx.Option, regularOptions []fx.Option, cliOptions ...CliOptions) {
	rootCmd.Use = appName

	// To add more cmd. Declare the cmd as a variable similar to rootCmd. And add it to rootCmd here.
	for _, f := range cliOptions {
		f(rootCmd)
	}

	// Configure Run function
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		// make a copy of cli exec context
		execCtx := cliCtx
		execCtx.Cmd = cmd
		execCtx.Args = args

		app := bootstrapper().NewApp(&execCtx, priorityOptions, regularOptions)
		app.Run()
	}
}
