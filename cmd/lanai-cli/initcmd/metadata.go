package initcmd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"path/filepath"
	"strings"
)

type ModuleMetadata struct {
	Module      *cmdutils.GoModule     `json:"-"`
	Executables map[string]*Executable `json:"execs"`
	Generates   []*Generate            `json:"generates"`
	Resources   []*Resource            `json:"resources"`
}

type Executable struct {
	Main string `json:"main"`
	Port int    `json:"port"`
}

type Resource struct {
	Pattern string `json:"pattern"`
	Output  string `json:"output"`
}

type Generate struct {
	Path string `json:"path"`
}

func validateModuleMetadata(ctx context.Context) error {
	if Module.Module = cmdutils.ResolveTargetModule(ctx); Module.Module == nil {
		return fmt.Errorf("unable to resolve module name in %s", cmdutils.GlobalArgs.WorkingDir)
	}

	// fix Executable.Main
	for k, v := range Module.Executables {
		fixed, e := fixPkgPath(ctx, v.Main, Module.Module.Path)
		if e != nil {
			return fmt.Errorf("invalid value of execs.%s.main: %v", k, e)
		}
		if fixed != v.Main {
			logger.WithContext(ctx).Debugf("Rewrite Main Path: %s => %s", v.Main, fixed)
		}
		v.Main = fixed
	}

	// fix Generates
	for i, v := range Module.Generates {
		fixed, e := fixPkgPath(ctx, v.Path, Module.Module.Path)
		if e != nil {
			return fmt.Errorf("invalid value of generates[%d].path: %v", i, e)
		}
		if fixed != v.Path {
			logger.WithContext(ctx).Debugf("Rewrite Generate Path: %s => %s", v.Path, fixed)
		}
		v.Path = fixed
	}

	// TODO more validation
	return nil
}

// fixPkgPath attempts to fix given pkg path if it's relative to go.mod folder
func fixPkgPath(ctx context.Context, path string, module string) (pkgPath string, err error) {
	pkgPath = path
	switch {
	case strings.ToLower(filepath.Ext(pkgPath)) == ".go":
		// go file should not prepend with module name
		return
	case strings.HasPrefix(strings.ToLower(pkgPath), strings.ToLower(module)):
		// already prepended with module name
		return
	}

	if ok, e := cmdutils.IsLocalPackageExists(ctx, pkgPath); e == nil && ok {
		return
	}

	// try add module string and check again
	pkgPath = module + "/" + path
	if ok, e := cmdutils.IsLocalPackageExists(ctx, pkgPath); e != nil {
		return "", e
	} else if !ok {
		return "", fmt.Errorf("unable to fix package path value [%s]. package [%s] doesn't exists", path, pkgPath)
	} else {
		return
	}
}
