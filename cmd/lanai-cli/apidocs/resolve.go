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

package apidocs

import (
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "github.com/spf13/cobra"
)

/**********************
	Command
 **********************/

var (
	ResolveCmd = &cobra.Command{
		Use:                "resolve <space_delimited_source_files>",
		Short:              "Generate OAS3 document from contract definitions and external references",
		Example:            `lanai-cli apidocs resolve -O configs/api-docs.yaml -T $GITHUB_TOKEN -R https://api.swaggerhub.com/domains/<organization>/<project>/8=>github://github.com/raw/<organization>/<project>/<tag/branch>/common-domain-8.yaml contracts/service-8.yaml contracts/service-1.json`,
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		Args:               ValidatePositionalArgs,
		RunE:               RunMerge,
	}
	ResolveArgs = ResolveArguments{
		Output:     "api-docs-merged.yaml",
		KeepExtRef: false,
	}
)

type ResolveArguments struct {
	Output            string   `flag:"output-file,O" desc:"Path of output file, relative to working directory"`
	KeepExtRef        bool     `flag:"keep-external-ref,k" desc:"Keep external $ref as-is (skip resolving external $ref)"`
	ConfigPath        string   `flag:"config,c" desc:"Path of configuration YAML, relative to working directory. \nThis is an alternative way of providing GitHub Token and external source replacement. \nNote: the value of GitHub Token can be an environment variable $EVN_VAR."`
	GitHubPATs        []string `flag:"github-token,T" desc:"GitHub's Personal Access Token(PAT). \nFormat: \"<token>[@<hostname>]\". \nWhen <hostname> is not specified, the token is used as default."`
	ReplaceExtSources []string `flag:"replace-external-source,R" desc:"Replace external reference sources with an alternative location. \nFormat: \"<original_url>=>[replaced_loc]\". \nSupported <replaced_loc> are 'http://', 'https://', 'github://' and local file. \ne.g. https://api.swaggerhub.com/domains/<organization>/<project>/8=>github://github.com/raw/<organization>/<project>/<tag/branch>/common-domain-8.yaml"`
}

func init() {
	cmdutils.PersistentFlags(ResolveCmd, &ResolveArgs)
}

func ValidatePositionalArgs(_ *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("missing source files, at least 1 source file")
	}
	return nil
}

func RunMerge(cmd *cobra.Command, args []string) error {

	if e := tryLoadConfig(); e != nil {
		return e
	}

	logger.Infof("Loading API documents...")
	docs, e := loadApiDocs(cmd.Context(), args)
	if e != nil {
		return e
	}

	if !ResolveArgs.KeepExtRef {
		logger.Infof("Resolving external $ref...")
		docs, e = tryResolveExtRefs(cmd.Context(), docs)
		if e != nil {
			return e
		}
	}

	logger.Infof("Merging...")
	merged, e := merge(cmd.Context(), docs)
	if e != nil {
		return e
	}

	logger.Infof("Writing to [%s]...", ResolveArgs.Output)
	return writeMergedToFile(cmd.Context(), merged)
}

/************************
	Config
 ************************/

var (
	ResolveConf = ResolveConfig{}
)

type ResolveConfig struct {
	GitHubTokens      []ResolveGitHubTokenMapping `json:"github-token"`
	ReplaceExtSources []ResolveExtSourceMapping   `json:"replace"`
}

type ResolveGitHubTokenMapping struct {
	Host  string `json:"host"`
	Token string `json:"token"`
}

type ResolveExtSourceMapping struct {
	Url string `json:"url"`
	To  string `json:"to"`
}

func tryLoadConfig() error {
	if ResolveArgs.ConfigPath == "" {
		return nil
	}
	if e := cmdutils.LoadYamlConfig(&ResolveConf, ResolveArgs.ConfigPath); e != nil {
		return fmt.Errorf(`invalid config file [%s]: %v`, ResolveArgs.ConfigPath, e)
	}
	return nil
}
