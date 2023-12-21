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
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

var (
	DropReplaceCmd = &cobra.Command{
		Use:                "drop-replace",
		Short:              "drop the replace directive for a given module and update go.sum",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunDropReplace,
	}
)

func RunDropReplace(cmd *cobra.Command, _ []string) error {
	toBeReplaced := utils.NewStringSet()

	for _, arg := range Args.Modules {
		m := strings.Split(arg, "@") //because arg can be module or module:version. if it's the latter, we take the part before :
		if len(m) != 1 && len(m) != 2 {
			return errors.New("input should be a comma separated list of module or module@version strings")
		}
		toBeReplaced.Add(m[0])
	}

	mod, err := cmdutils.GetGoMod(cmd.Context())
	if err != nil {
		return err
	}

	changed := false
	for _, replace := range mod.Replace {
		logger.Infof("found replace for %s, %s", replace.Old.Path, replace.Old.Version)

		//we only drop the replace for the module whose dependency we updated
		//there may be replace that are pointing to other module (not local)
		if toBeReplaced.Has(replace.Old.Path) {
			err = cmdutils.DropReplace(cmd.Context(), replace.Old.Path, replace.Old.Version)
			if err != nil {
				return err
			}
			changed = true
		}
	}

	// just in case, drop invalid replace
	dropped, e := cmdutils.DropInvalidReplace(cmd.Context())
	if e != nil {
		return fmt.Errorf("failed to drop invalid replace: %v", e)
	}
	changed = len(dropped) != 0

	if changed {
		if e := cmdutils.GoModTidy(cmd.Context(), nil); e != nil {
			return e
		}
	}

	// mark changes if requested
	msg := fmt.Sprintf("dropped replaces in go.mod for CI/CD process")
	tag, e := markChangesIfRequired(cmd.Context(), msg, cmdutils.GitFilePattern("go.mod", "go.sum"))
	if e == nil && tag != "" {
		logger.WithContext(cmd.Context()).Infof(`Modified go.mod/go.sum are tagged with Git Tag [%s]`, tag)
	}
	return e
}
