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

package dev

import (
    "context"
    "fmt"
    "github.com/bmatcuk/doublestar/v4"
    "github.com/cisco-open/go-lanai/cmd/lanai-cli/cmdutils"
    "github.com/cisco-open/go-lanai/pkg/utils"
    "path/filepath"
    "strings"
)

// resolveLocalMods search for given search paths and find all go.mod files
func findLocalGoModFiles(searchPaths ...string) ([]string, error) {
	var ret []string
	for _, path := range searchPaths {
		relPath := toRelativePath(path, cmdutils.GlobalArgs.WorkingDir)
		if len(relPath) == 0 {
			logger.Warnf(`Search path "%s" is ignored`, path)
			continue
		}
		modPaths, e := cmdutils.LookupFiles("/**/go.mod", relPath)
		if e != nil {
			return nil, fmt.Errorf(`unable to find go.mod in "%s": %v`, path, e)
		}
		ret = append(ret, modPaths...)
	}
	return ret, nil
}

// resolveLocalMods search for given search paths and find and parse all go.mod files.
// It returns a map of Module name with its absolute file path
func resolveLocalMods(ctx context.Context, searchPaths ...string) (map[string]string, error) {
	ret := map[string]string{}
	modPaths, e := findLocalGoModFiles(searchPaths...)
	if e != nil {
		return nil, e
	}

	for _, modPath := range modPaths {
		mod, e := cmdutils.GetGoMod(ctx, cmdutils.GoCmdModFile(modPath))
		if e != nil {
			logger.Warnf(`Ignoring "%s" due to error: %v`, modPath, e)
			continue
		}
		ret[mod.Module.Path] = modPath
	}
	return ret, nil
}

func toRelativePath(path string, base string) string {
	if !filepath.IsAbs(path) {
		return path
	}
	path, e := filepath.Rel(base, path)
	if e != nil {
		return ""
	}
	return path
}

func pathMatches(path string, patterns utils.StringSet) bool {
	for pattern := range patterns {
		if ok, e := doublestar.PathMatch(pattern, path); e == nil && ok {
			return true
		}
	}
	return false
}

func resolveLocalReplacePath(modPath string, base string) string {
	relModPath := filepath.Dir(toRelativePath(modPath, base))
	switch {
	case !strings.HasPrefix(relModPath, "."):
		// mod path is a sub folder of current folder
		relModPath = "./" + relModPath
	case relModPath == "..":
		relModPath = "../"
	case relModPath == ".":
		relModPath = "./"
	}
	return relModPath
}