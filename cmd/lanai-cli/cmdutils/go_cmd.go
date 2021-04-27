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

var (
	targetTmpGoModFile string
	targetModule       *GoModule
	targetModuleOnce   = sync.Once{}

	packageImportPathCache     map[string]*GoPackage
	packageImportPathCacheOnce = sync.Once{}
)

func ResolveTargetModule(ctx context.Context) *GoModule {
	targetModuleOnce.Do(func() {
		// first, prepare a mod file to read
		modFile, e := prepareTargetGoModFile(ctx)
		if e != nil {
			logger.WithContext(ctx).Errorf("unable to prepare temporary go.mod file to resolve target module: %v", e)
		}
		targetTmpGoModFile = modFile

		// find module
		mods, e := FindModule(ctx, modFile)
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
		packageImportPathCache, err = FindPackages(ctx, targetTmpGoModFile, module.Path)
		if err != nil {
			logger.WithContext(ctx).Errorf("unable to resolve local packages in module %s", module.Path)
		}
	})
	return packageImportPathCache
}

func FindModule(ctx context.Context, modFile string, modules ...string) ([]*GoModule, error) {
	var modFileArg string
	if modFile != "" {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile)
	}
	result, e := GoCommandDecodeJson(ctx, &GoModule{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go list -m -json%s %s", modFileArg, strings.Join(modules, " "))),
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

func FindPackages(ctx context.Context, modFile string, modules ...string) (map[string]*GoPackage, error) {
	var modFileArg string
	if modFile != "" {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile)
	}
	result, e := GoCommandDecodeJson(ctx, &GoPackage{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go list -json%s %s/...", modFileArg, strings.Join(modules, " "))),
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

func GoCommandDecodeJson(ctx context.Context, model interface{}, opts ...ShCmdOptions) (ret []interface{}, err error) {
	mt := reflect.TypeOf(model)
	if mt.Kind() == reflect.Ptr {
		mt = mt.Elem()
	}

	pr, pw := io.Pipe()
	opts = append(opts, ShellStdOut(pw))
	ech := make(chan error, 1)
	go func() {
		defer pw.Close()
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

func IsLocalPackageExists(ctx context.Context, pkgPath string) (bool, error) {
	cache := PackageImportPathCache(ctx)
	if cache == nil {
		return false, fmt.Errorf("package import path cache is not available")
	}
	_, ok := cache[pkgPath]
	return ok, nil
}

func DropReplace(ctx context.Context, module string, version string, modFile ...string) error {
	var modFileArg string
	if len(modFile) != 0 {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile[0])
	}

	var cmd ShCmdOptions
	if version == "" {
		cmd = ShellCmd(fmt.Sprintf("go mod edit%s -dropreplace %s", modFileArg, module))
	} else {
		cmd = ShellCmd(fmt.Sprintf("go mod edit%s -dropreplace %s@%s", modFileArg, module, version))
	}

	logger.Infof("dropping replace directive %s, %s", module, version)

	_, err := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		cmd,
		ShellStdOut(os.Stdout))

	return err
}

func DropRequire(ctx context.Context, module string, modFile ...string) error {
	var modFileArg string
	if len(modFile) != 0 {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile[0])
	}

	logger.Infof("dropping require directive %s", module)

	_, err := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go mod edit%s -droprequire %s", modFileArg, module)),
		ShellStdOut(os.Stdout))

	return err
}

func GoGet(ctx context.Context, module string, versionQuery string) error {
	_, e := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go get %s@%s", module, versionQuery)),
		ShellStdOut(os.Stdout))
	return e
}

func GoModTidy(ctx context.Context, modFile ...string) error {
	var modFileArg string
	if len(modFile) != 0 {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile[0])
	}
	_, e := RunShellCommands(ctx,
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go mod tidy%s", modFileArg)),
		ShellStdOut(os.Stdout))
	return e
}

func GetGoMod(ctx context.Context, modFile ...string) (*GoMod, error){
	var modFileArg string
	if len(modFile) != 0 {
		modFileArg = fmt.Sprintf(" -modfile %s", modFile[0])
	}
	result, e := GoCommandDecodeJson(ctx, &GoMod{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go mod edit -json%s", modFileArg)),
	)
	if e != nil {
		return nil, e
	}

	m := result[0].(*GoMod)
	return m, nil
}

func tmpGoModFile() string {
	return GlobalArgs.AbsPath(GlobalArgs.TmpDir, "go.tmp.mod")
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

	// read mod file
	mod, e := GetGoMod(ctx, tmpModFile)
	if e != nil {
		return "", fmt.Errorf("error when read go.mod: %v", e)
	}

	// try drop replace for non-exiting local reference in tmp folder
	for _, v := range mod.Replace {
		replaced := v.New.Path
		// we only care if the replaced path start with "/" or ".",
		// i.e. we will ignore url path such as "cto-github.cisco.com/NFV-BU/go-lanai"
		if replaced == "" || !filepath.IsAbs(replaced) && !strings.HasPrefix(replaced, ".") {
			continue
		}

		if !filepath.IsAbs(replaced) {
			replaced = filepath.Clean(GlobalArgs.WorkingDir + "/" + replaced)
		}
		if !isFileExists(replaced) {
			if e := DropReplace(ctx, v.Old.Path, v.Old.Version, tmpModFile); e != nil {
				return "", fmt.Errorf("error when dropping replace %s: %v", v.Old.Path, e)
			}
			if e := DropRequire(ctx, v.Old.Path, tmpModFile); e != nil {
				return "", fmt.Errorf("error when dropping require %s: %v", v.Old.Path, e)
			}
		}
	}
	return tmpModFile, nil
}