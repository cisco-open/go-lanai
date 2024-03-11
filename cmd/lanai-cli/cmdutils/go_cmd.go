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
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
)

const (
	goCmdModEdit = `go mod edit`
)

var (
	targetTmpGoModFile string
	targetModule       *GoModule
	targetModuleOnce   = sync.Once{}

	packageImportPathCache     map[string]*GoPackage
	packageImportPathCacheOnce = sync.Once{}
)

type GoCmdOptions func(goCmd *string)

func GoCmdModFile(modFile string) GoCmdOptions {
	return func(goCmd *string) {
		if modFile == "" {
			return
		}
		*goCmd = fmt.Sprintf("%s -modfile %s", *goCmd, modFile)
	}
}

func ResolveTargetModule(ctx context.Context) *GoModule {
	targetModuleOnce.Do(func() {
		// first, prepare a mod file to read
		modFile, e := prepareTargetGoModFile(ctx)
		if e != nil {
			logger.WithContext(ctx).Errorf("unable to prepare temporary go.mod file to resolve target module: %v", e)
		}
		targetTmpGoModFile = modFile

		// find module
		mods, e := FindModule(ctx, []GoCmdOptions{GoCmdModFile(modFile)})
		if e == nil && len(mods) == 1 {
			targetModule = mods[0]
		} else if e != nil {
			logger.WithContext(ctx).Errorf("unable to resolve target module name: %v", e)
		} else {
			logger.WithContext(ctx).Errorf("resolved multiple modules in working directory")
		}
	})
	return targetModule
}

func PackageImportPathCache(ctx context.Context) map[string]*GoPackage {
	packageImportPathCacheOnce.Do(func() {
		module := ResolveTargetModule(ctx)
		if module == nil {
			return
		}
		var err error
		packageImportPathCache, err = FindPackages(ctx, []GoCmdOptions{GoCmdModFile(targetTmpGoModFile)}, module.Path)
		if err != nil {
			logger.WithContext(ctx).Errorf("unable to resolve local packages in module %s", module.Path)
		}
	})
	return packageImportPathCache
}

func FindModule(ctx context.Context, opts []GoCmdOptions, modules ...string) ([]*GoModule, error) {
	cmd := "go list -m -json"
	for _, f := range opts {
		f(&cmd)
	}
	cmd = fmt.Sprintf("%s %s", cmd, strings.Join(modules, " "))

	result, e := GoCommandDecodeJson(ctx, &GoModule{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(cmd),
	)
	if e != nil {
		return nil, e
	}

	var ret []*GoModule
	for _, v := range result {
		m := v.(*GoModule)
		ret = append(ret, m)
	}
	return ret, nil
}

func FindPackages(ctx context.Context, opts []GoCmdOptions, modules ...string) (map[string]*GoPackage, error) {
	cmd := "go list -json -find"
	for _, f := range opts {
		f(&cmd)
	}
	cmd = fmt.Sprintf("%s %s/...", cmd, strings.Join(modules, " "))

	result, e := GoCommandDecodeJson(ctx, &GoPackage{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(cmd),
	)
	if e != nil {
		return nil, e
	}

	pkgs := map[string]*GoPackage{}
	for _, v := range result {
		pkg := v.(*GoPackage)
		pkgs[pkg.ImportPath] = pkg
	}
	return pkgs, nil
}

// DropInvalidReplace go through the go.mod file and find replace directives that point to a non-existing local directory
func DropInvalidReplace(ctx context.Context, opts ...GoCmdOptions) (ret []*Replace, err error) {
	mod, e := GetGoMod(ctx, opts...)
	if e != nil {
		return nil, e
	}

	cmdOpts := []ShCmdOptions{
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellStdOut(os.Stdout),
	}
	for _, v := range mod.Replace {
		if isInvalidReplace(&v) {
			ret = append(ret, &v)
			cmdOpts = append(cmdOpts, dropReplaceCmd(v.Old.Path, v.Old.Version, opts))
		}
	}
	if len(ret) == 0 {
		return
	}

	if _, e := RunShellCommands(ctx, cmdOpts...); e != nil {
		return nil, e
	}
	return
}

// RestoreInvalidReplace works together with DropInvalidReplace
func RestoreInvalidReplace(ctx context.Context, replaces []*Replace, opts ...GoCmdOptions) error {
	if len(replaces) == 0 {
		return nil
	}

	cmdOpts := []ShCmdOptions{
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellStdOut(os.Stdout),
	}
	for _, v := range replaces {
		cmdOpts = append(cmdOpts, setReplaceCmd([]*Replace{v}, opts))
	}

	_, err := RunShellCommands(ctx, cmdOpts...)

	return err
}

// SetReplace Set given replaces in go.mod
func SetReplace(ctx context.Context, replaces []*Replace, opts ...GoCmdOptions) error {
	if len(replaces) == 0 {
		return nil
	}

	cmdOpts := []ShCmdOptions{
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellStdOut(os.Stdout),
		setReplaceCmd(replaces, opts),
	}
	_, err := RunShellCommands(ctx, cmdOpts...)
	return err
}

func DropReplace(ctx context.Context, module string, version string, opts ...GoCmdOptions) error {
	logger.Infof("dropping replace directive %s, %s", module, version)
	_, err := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		dropReplaceCmd(module, version, opts),
		ShellStdOut(os.Stdout))

	return err
}

func DropRequire(ctx context.Context, module string, opts ...GoCmdOptions) error {
	cmd := goCmdModEdit
	for _, f := range opts {
		f(&cmd)
	}
	cmd = fmt.Sprintf("%s -droprequire %s", cmd, module)

	logger.Infof("dropping require directive %s", module)
	_, err := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(cmd),
		ShellStdOut(os.Stdout))

	return err
}

func GoGet(ctx context.Context, module string, versionQuery string, opts ...GoCmdOptions) error {
	cmd := "go get"
	for _, f := range opts {
		f(&cmd)
	}
	cmd = fmt.Sprintf("%s %s@%s", cmd, module, versionQuery)

	_, e := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellStdOut(os.Stdout),
		ShellCmd(cmd),
	)
	return e
}

func GoModTidy(ctx context.Context, extraShellOptions []ShCmdOptions, opts ...GoCmdOptions) error {
	cmd := "go mod tidy"
	for _, f := range opts {
		f(&cmd)
	}

	shellOptions := []ShCmdOptions{
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(cmd),
		ShellStdOut(os.Stdout),
	}
	shellOptions = append(shellOptions, extraShellOptions...)
	_, e := RunShellCommands(ctx, shellOptions...)
	return e
}

func GetGoMod(ctx context.Context, opts ...GoCmdOptions) (*GoMod, error) {
	cmd := fmt.Sprintf("go mod edit -json")
	for _, f := range opts {
		f(&cmd)
	}
	result, e := GoCommandDecodeJson(ctx, &GoMod{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(cmd),
	)
	if e != nil {
		return nil, e
	}

	m := result[0].(*GoMod)
	return m, nil
}

/***********************
	Exported Helpers
 ***********************/

func IsLocalPackageExists(ctx context.Context, pkgPath string) (bool, error) {
	cache := PackageImportPathCache(ctx)
	if cache == nil {
		return false, fmt.Errorf("package import path cache is not available")
	}
	_, ok := cache[pkgPath]
	return ok, nil
}

func GoCommandDecodeJson(ctx context.Context, model interface{}, opts ...ShCmdOptions) (ret []interface{}, err error) {
	mt := reflect.TypeOf(model)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
	}

	pr, pw := io.Pipe()
	opts = append(opts, ShellStdOut(pw))
	ech := make(chan error, 1)
	go func() {
		defer func() { _ = pw.Close() }()
		defer close(ech)
		_, e := RunShellCommands(ctx, opts...)
		if e != nil {
			ech <- e
		}
	}()

	dec := json.NewDecoder(pr)
	for {
		m := reflect.New(mt).Interface()
		if e := dec.Decode(&m); e != nil {
			if e != io.EOF {
				err = e
			}
			break
		}
		ret = append(ret, m)
	}

	if e := <-ech; e != nil {
		err = e
		return
	}
	return
}

/***********************
	Helper Functions
 ***********************/

func withVersionQuery(module string, version string) string {
	if version == "" {
		return module
	}

	return fmt.Sprintf("%s@%s", module, version)
}

func dropReplaceCmd(module string, version string, opts []GoCmdOptions) ShCmdOptions {
	cmd := goCmdModEdit
	for _, f := range opts {
		f(&cmd)
	}
	cmd = fmt.Sprintf("%s -dropreplace %s", cmd, withVersionQuery(module, version))

	return ShellCmd(cmd)
}

func setReplaceCmd(replaces []*Replace, opts []GoCmdOptions) ShCmdOptions {
	cmd := goCmdModEdit
	for _, f := range opts {
		f(&cmd)
	}
	cmds := []string{cmd}
	for _, replace := range replaces {
		from := withVersionQuery(replace.Old.Path, replace.Old.Version)
		to := withVersionQuery(replace.New.Path, replace.New.Version)
		cmds = append(cmds, fmt.Sprintf("-replace %s=%s", from, to))
	}

	return ShellCmd(strings.Join(cmds, " "))
}

func tmpGoModFile() string {
	return GlobalArgs.AbsPath(GlobalArgs.TmpDir, "go.tmp.mod")
}

func isInvalidReplace(replace *Replace) bool {
	replaced := replace.New.Path
	// we only care if the replaced path start with "/" or ".",
	// i.e. we will ignore url path such as "github.com/cisco-open/go-lanai"
	if replaced == "" || !filepath.IsAbs(replaced) && !strings.HasPrefix(replaced, ".") {
		return false
	}

	if !filepath.IsAbs(replaced) {
		replaced = filepath.Clean(GlobalArgs.WorkingDir + "/" + replaced)
	}
	return !isFileExists(replaced)
}

func prepareTargetGoModFile(ctx context.Context) (string, error) {
	tmpModFile := tmpGoModFile()
	// make a copy of go.mod and go.sum in tmp folder
	files := map[string]string{
		"go.mod": tmpModFile,
		"go.sum": GlobalArgs.AbsPath(GlobalArgs.TmpDir, "go.tmp.sum"),
	}
	if e := copyFiles(ctx, files); e != nil {
		return "", fmt.Errorf("error when copying go.mod: %v", e)
	}

	// drop invalid replace
	replaces, e := DropInvalidReplace(ctx, GoCmdModFile(tmpModFile))
	if e != nil {
		return "", fmt.Errorf("error when drop invalid replaces: %v", e)
	}

	// drop require as well
	// This is because when there is a "replace" directive in the go.mod, the go.sum
	// file usually doesn't include the entries that were "replaced" by local copy.
	// i.e. running "go mod tidy" on such a go.mod file will result in the replaced entry
	// being removed from go.sum.
	// When the go.sum is in this state, the result of the "go list" command depends on
	// if the "replace" directive points to a valid local directory.
	// If the local directory doesn't exist, the command will result in error.
	// If the non-existent "replace" directive is dropped, but the "require" directive is not dropped,
	// we would get an error complaining about the missing go.sum entry.
	// Therefor we drop the "require" directive as well.
	// This is ok because how we intend to use the resulting go.mod file.
	// We will use this go.mod to find packages in the current service. So its dependencies are not important.
	// However, we need to be careful to not use go commands that resolves dependencies.
	// For example, in FindPackages we use the "-find" tag in the "go list -json -find" to not resolve dependencies.
	for _, v := range replaces {
		if e := DropRequire(ctx, v.Old.Path, GoCmdModFile(tmpModFile)); e != nil {
			return "", fmt.Errorf("error when dropping require %s: %v", v.Old.Path, e)
		}
	}

	return tmpModFile, nil
}
