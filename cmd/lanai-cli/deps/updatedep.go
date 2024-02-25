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

package deps

import (
    "errors"
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "github.com/spf13/cobra"
    "strings"
)

var (
	UpdateDepCmd = &cobra.Command{
		Use:                "update",
		Short:              "Update module dependencies with given branches and update go.sum",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunUpdateDep,
	}
)

func RunUpdateDep(cmd *cobra.Command, _ []string) error {
	//process input args to see which module's dependency needs to be updated
	moduleToBranch := make(map[string]string)
	for _, arg := range Args.Modules {
		mbr := strings.Split(arg, "@")
		if len(mbr) != 2 {
			logger.Errorf("%s doesn't follow the module@branch format", mbr)
			return errors.New("can't parse module path")
		}
		m := mbr[0]
		br := mbr[1]

		moduleToBranch[m] = br
	}

	dropped, e := cmdutils.DropInvalidReplace(cmd.Context())
	if e != nil {
		return fmt.Errorf("unable to temporarily drop invalid 'replace' in go.mod: %v", e)
	}

	// update their dependencies
	for module, branch := range moduleToBranch {
		logger.Infof("processing %s@%s", module, branch)
		e := cmdutils.GoGet(cmd.Context(), module, branch)
		if e != nil {
			return nil
		}
	}

	// go mod tidy to update implicit dependencies changes
	if e := cmdutils.GoModTidy(cmd.Context(), nil); e != nil {
		return e
	}

	if e := cmdutils.RestoreInvalidReplace(cmd.Context(), dropped); e != nil {
		return fmt.Errorf("unable to restore temporarily dropped 'replace' in go.mod: %v", e)
	}

	// mark changes if requested
	msg := fmt.Sprintf("updated versions of private modules")
	tag, e := markChangesIfRequired(cmd.Context(), msg, cmdutils.GitFilePattern("go.mod", "go.sum"))
	if e == nil && tag != "" {
		logger.WithContext(cmd.Context()).Infof(`Modified go.mod/go.sum are tagged with Git Tag [%s]`, tag)
	}
	return e
}
