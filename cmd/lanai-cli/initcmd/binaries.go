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
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
)

var defaultBinaries = map[string]string{
	"github.com/axw/gocov/gocov":  "v1.1.0",
	"github.com/AlekSi/gocov-xml": "v1.1.0",
	"gotest.tools/gotestsum":      "v1.13.0",
	//"github.com/golangci/golangci-lint/cmd/golangci-lint": "v1.64.8",
	"github.com/golangci/golangci-lint/v2/cmd/golangci-lint": "v2.12.2",
}

var exclusiveBinaries = [][]string{
	{"github.com/golangci/golangci-lint/**"},
}

func resolveBinariesFromCache(ctx context.Context) (map[string]string, error) {
	if binaries, ok := Module.Data[kDataBinaries].(map[string]string); ok {
		return binaries, nil
	}
	return resolveBinaries(ctx)
}

func resolveBinaries(_ context.Context) (map[string]string, error) {
	binaries := make(map[string]string)

	for p, v := range defaultBinaries {
		binaries[p] = v
	}
	for _, b := range Module.Binaries {
		switch {
		case len(b.Package) == 0:
			logger.Warnf(`Invalid binaries entry in Module.yml: package="%s", version="%s"`, b.Package, b.Version)
		case len(b.Version) == 0:
			if _, ok := binaries[b.Package]; ok {
				logger.Warnf(`Skipping default binaries install: %s@%s`, b.Package, binaries[b.Package])
				delete(binaries, b.Package)
			}
		default:
			binaries[b.Package] = b.Version
		}
	}

	if conflicts := findConflictingBinaries(binaries); len(conflicts) != 0 {
		for _, conflict := range conflicts {
			logger.Errorf("Detected conflicting binaries in Module.yml. Following tools should not be installed together:\n    - %s", strings.Join(conflict, "\n      "))
		}
		return nil, fmt.Errorf("detected %d group(s) of conflicting binaries. please check Module.yml", len(conflicts))
	}
	return binaries, nil
}

func installBinaries(ctx context.Context) error {

	opts := []cmdutils.ShCmdOptions{cmdutils.ShellShowCmd(true), cmdutils.ShellUseWorkingDir(), cmdutils.ShellStdOut(os.Stdout)}

	binaries, e := resolveBinariesFromCache(ctx)
	if e != nil {
		return e
	}

	for p, v := range binaries {
		installCmd := fmt.Sprintf("go install %s@%s", p, v)
		opts = append(opts, cmdutils.ShellCmd(installCmd))
	}

	_, e = cmdutils.RunShellCommands(ctx, opts...)
	return e
}

func findConflictingBinaries(binaries map[string]string) [][]string {
	var conflicts [][]string
	for _, patterns := range exclusiveBinaries {
		found := make([]string, 0, 2)
		for pkg := range binaries {
			if patternsMatchString(patterns, pkg) {
				found = append(found, pkg)
			}
		}
		if len(found) > 1 {
			conflicts = append(conflicts, found)
		}
	}
	if len(conflicts) > 0 {
		return conflicts
	}
	return nil
}

func patternsMatchString(patterns []string, str string) bool {
	for _, pattern := range patterns {
		if ok, e := doublestar.Match(pattern, str); e == nil && ok {
			return true
		}
	}
	return false
}
