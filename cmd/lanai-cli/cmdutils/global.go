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

package cmdutils

import (
	"github.com/cisco-open/go-lanai/pkg/log"
	"os"
	"path"
	"path/filepath"
)

const (
	ModulePath = "github.com/cisco-open/go-lanai"
)

var (
	logger     = log.New("Build")
	GlobalArgs = Global{
		WorkingDir: DefaultWorkingDir(),
		TmpDir:     DefaultTemporaryDir(),
		OutputDir:  DefaultOutputDir(),
	}
)

type Global struct {
	WorkingDir string `flag:"workspace,w" desc:"working directory containing 'go.mod'. All non-absolute paths are relative to this directory"`
	TmpDir     string `flag:"tmp-dir" desc:"temporary directory."`
	OutputDir  string `flag:"output,o" desc:"output directory. All non-absolute paths for output are relative to this directory"`
	Verbose    bool   `flag:"debug" desc:"show debug information"`
}

func (g Global) AbsPath(base, path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Clean(base + "/" + path)
}

func DefaultWorkingDir() string {
	return goModDir()
}

func DefaultTemporaryDir() string {
	const relative = ".tmp"
	return PathRelativeToModuleDir(relative)
}

func DefaultOutputDir() string {
	const relative = "dist"
	return PathRelativeToModuleDir(relative)
}

func PathRelativeToModuleDir(relativePath string) string {
	if path.IsAbs(relativePath) {
		panic("PathRelativeToModuleDir only takes relative path")
	}
	base := goModDir()
	if base == "" {
		return DefaultWorkingDir() + "/" + relativePath
	}
	return base + "/" + relativePath
}

func currentDir() string {
	currDir, e := os.Getwd()
	if e != nil {
		panic(e)
	}
	return currDir
}

// goModDir works from current directory backward along the FS tree until find go.mod file
// if root directory is hit and it's still not found, return the currentDir
func goModDir() string {
	currDir, e := os.Getwd()
	if e != nil {
		panic(e)
	}

	for dir := currDir; dir != "" && dir != "."; dir = filepath.Dir(dir) {
		gomodPath := dir + "/go.mod"
		if _, e := os.Stat(gomodPath); e == nil {
			return dir
		}
		if dir == "/" {
			break
		}
	}
	return currentDir()
}
