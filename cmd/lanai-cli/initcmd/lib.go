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
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/spf13/cobra"
)

var (
	LibInitCmd    = &cobra.Command{
		Use:                InitLibsName,
		Short:              "Initialize library module, generating additional Makefile rules, etc.",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunLibsInit,
	}
)

func RunLibsInit(cmd *cobra.Command, _ []string) error {
	if e := cmdutils.LoadYamlConfig(&Module, Args.Metadata); e != nil {
		return e
	}

	if e := validateModuleMetadata(cmd.Context()); e != nil {
		return e
	}

	if e := generateLibsCICDMakefile(cmd.Context()); e != nil {
		return e
	}

	if e := installBinaries(cmd.Context()); e != nil {
		return e
	}

	return nil
}
