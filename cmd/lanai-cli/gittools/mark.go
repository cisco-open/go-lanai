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
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
	"github.com/spf13/cobra"
)

var (
	MarkCmd = &cobra.Command{
		Use:                "mark",
		Short:              "tools to create/revert/retag commit marks using a lightweight tag",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
	}
	MarkArgs = MarkArguments{}
)

type MarkArguments struct {
	MarkTag string   `flag:"tag,t,required" desc:"Local git tag to mark the worktree. Scripts can go back to the result of this step by using the same tag"`
}

func init() {
	MarkCmd.AddCommand(MarkCreateCmd)
	MarkCmd.AddCommand(MarkRevertCmd)
	MarkCmd.AddCommand(MarkReTagCmd)
	cmdutils.PersistentFlags(MarkCmd, &MarkArgs)
}

