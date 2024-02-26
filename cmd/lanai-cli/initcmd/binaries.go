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
	"context"
	"fmt"
	"os"

	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
)

var defaultBinaries = map[string]string{
	"github.com/axw/gocov/gocov":                          "v1.1.0",
	"github.com/AlekSi/gocov-xml":                         "v1.0.0",
	"gotest.tools/gotestsum":                              "v1.8.0",
	"github.com/golangci/golangci-lint/cmd/golangci-lint": "v1.55.2",
	"github.com/jstemmer/go-junit-report":                 "v0.9.1",
}

func installBinaries(ctx context.Context) error {
	opts := []cmdutils.ShCmdOptions{cmdutils.ShellShowCmd(true), cmdutils.ShellUseWorkingDir(), cmdutils.ShellStdOut(os.Stdout)}

	binaries := make(map[string]string)

	for p, v := range defaultBinaries {
		binaries[p] = v
	}
	for _, b := range Module.Binaries {
		if b.Package == "" || b.Version == "" {
			logger.Warnf(`Invalid binaries entry in Module.yml: package="%s", version="%s"`, b.Package, b.Version)
			continue
		}
		binaries[b.Package] = b.Version
	}

	for p, v := range binaries {
		installCmd := fmt.Sprintf("go install %s@%s", p, v)
		opts = append(opts, cmdutils.ShellCmd(installCmd))
	}

	_, e := cmdutils.RunShellCommands(ctx, opts...)
	return e
}
