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
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Short: "A go-lanai based application.",
	Long: "This is a go-lanai based application.",
	FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func NewAppCmd(appName string, priorityOptions []fx.Option, regularOptions []fx.Option) {
	rootCmd.Use = appName
	app := newApp(rootCmd, priorityOptions, regularOptions)
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		app.Run()
	}

	//To add more cmd. Declare the cmd as a variable similar to rootCmd. And add it to rootCmd here.
}