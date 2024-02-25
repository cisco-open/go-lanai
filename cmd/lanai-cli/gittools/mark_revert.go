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

package gittools

import (
    "fmt"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "github.com/spf13/cobra"
    "strings"
)

var (
	MarkRevertCmd = &cobra.Command{
		Use:                "revert",
		Short:              "revert to the commit marked by given tag, without changing current branch",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunRevertMark,
	}
	MarkRevertArgs = MarkRevertArguments{}
)

type MarkRevertArguments struct {
	DiscardLocal    bool   `flag:"force-discard-changes,f" desc:"when set, local changes are discarded"`
}

func init() {
	cmdutils.PersistentFlags(MarkRevertCmd, &MarkRevertArgs)
}

func RunRevertMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	if tag == "" {
		return fmt.Errorf("tag is required flag and cannot be empty")
	}

	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	if e := gitutils.ResetToMarkedCommit(tag, MarkRevertArgs.DiscardLocal); e != nil {
		return fmt.Errorf("unable to revert to marked commit with tag %s: %v", tag, e)
	}

	logger.WithContext(cmd.Context()).Infof(`Reverted to commit marked as tag [%s]`, tag)
	return nil
}