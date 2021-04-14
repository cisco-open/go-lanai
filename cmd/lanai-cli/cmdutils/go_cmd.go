package cmdutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"sync"
)

var (
	targetModule               *GoModule
	targetModuleOnce           = sync.Once{}

	packageImportPathCache     map[string]*GoPackage
	packageImportPathCacheOnce = sync.Once{}
)

func ResolveTargetModule(ctx context.Context) *GoModule {
	targetModuleOnce.Do(func() {
		mods, e := FindModule(ctx)
		if e == nil && len(mods) == 1 {
			targetModule = mods[0]
		} else if e != nil {
			logger.WithContext(ctx).Errorf("unable to resolve target module name")
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
		packageImportPathCache, err = FindPackages(ctx, module.Path)
		if err != nil {
			logger.WithContext(ctx).Errorf("unable to resolve local packages in module", module.Path)
		}
	})
	return packageImportPathCache
}

func FindModule(ctx context.Context, modules ...string) ([]*GoModule, error) {
	result, e := GoCommandDecodeJson(ctx, &GoModule{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go list -m -json %s", strings.Join(modules, " "))),
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

func FindPackages(ctx context.Context, modules ...string) (map[string]*GoPackage, error) {
	result, e := GoCommandDecodeJson(ctx, &GoPackage{},
		ShellShowCmd(true),
		ShellUseWorkingDir(),
		ShellCmd(fmt.Sprintf("go list -json %s/...", strings.Join(modules, " "))),
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
