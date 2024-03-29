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
	MarkCreateCmd = &cobra.Command{
		Use:                "create",
		Short:              "Mark current commit with given tag name. optionally commit local changes before mark",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunCreateMark,
	}
	MarkCreateArgs = MarkCreateArguments{
		CommitMsg: "changes for git tagging",
	}
)

type MarkCreateArguments struct {
	Commit    bool   `flag:"include-local-changes" desc:"when set, local changes in worktree are commited before creating the mark"`
	CommitMsg string `flag:"commit-message,m" desc:"message used to commit local changes; ignored when 'include-local-changes' is not set. "`
}

func init() {
	cmdutils.PersistentFlags(MarkCreateCmd, &MarkCreateArgs)
}

func RunCreateMark(cmd *cobra.Command, _ []string) error {
	tag := strings.TrimSpace(MarkArgs.MarkTag)
	if tag == "" {
		return fmt.Errorf("tag is required flag and cannot be empty")
	}


	gitutils, e := cmdutils.NewGitUtilsWithWorkingDir()
	if e != nil {
		return e
	}
	gitutils = gitutils.WithContext(cmd.Context())

	// commit and mark
	if MarkCreateArgs.Commit {
		logger.WithContext(cmd.Context()).Infof(`All changes are committed and as tag [%s]`, tag)
		return gitutils.MarkWorktree(tag, MarkCreateArgs.CommitMsg, false)
	}

	// mark head
	hash, e := gitutils.HeadCommitHash()
	if e != nil {
		return fmt.Errorf("unable to get HEAD commit hash: %v", e)
	}

	if e := gitutils.MarkCommit(tag, hash); e != nil {
		return e
	}
	logger.WithContext(cmd.Context()).Infof(`Current HEAD is marked as tag [%s]`, tag)
	return nil
}
