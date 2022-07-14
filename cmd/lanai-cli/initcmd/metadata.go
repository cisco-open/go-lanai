package initcmd

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type ModuleMetadata struct {
	CliModPath  string                 `json:"-"`
	Module      *cmdutils.GoModule     `json:"-"`
	Name        string                 `json:"name"`
	Executables map[string]*Executable `json:"execs"`
	Resources   []*Resource            `json:"resources"`
	Generates   []*Generate            `json:"generates"`
	Binaries    []*Binary              `json:"binaries"`
	Sources     []*Source              `json:"sources"`
}

type Executable struct {
	Main  string `json:"main"`
	Port  int    `json:"port"`
	Ports []int  `json:"ports"`
	Type  string `json:"type"`
}

type Resource struct {
	Pattern string `json:"pattern"`
	Output  string `json:"output"`
}

type Generate struct {
	Path string `json:"path"`
}

type Binary struct {
	Package string `json:"package"`
	Version string `json:"version"`
}

type Source struct {
	Path string `json:"path"`
}

var (
	defaultSources = []*Source{
		{Path: "pkg"},
		{Path: "internal"},
	}
)

func validateModuleMetadata(ctx context.Context) error {
	if Module.Module = cmdutils.ResolveTargetModule(ctx); Module.Module == nil {
		return fmt.Errorf("unable to resolve module name in %s", cmdutils.GlobalArgs.WorkingDir)
	}

	// fix Executable
	for k, v := range Module.Executables {
		fixed, e := fixPkgPath(ctx, v.Main, Module.Module.Path)
		if e != nil {
			return fmt.Errorf("invalid value of execs.%s.main: %v", k, e)
		}
		if fixed != v.Main {
			logger.WithContext(ctx).Debugf("Rewrite Main Path: %s => %s", v.Main, fixed)
		}
		v.Main = fixed
		if v.Port > 0 {
			v.Ports = append([]int{v.Port}, v.Ports...)
		}
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

	// fix Sources
	var sources []*Source
	if len(Module.Sources) == 0 {
		Module.Sources = defaultSources
	}
	for i, v := range Module.Sources {
		src := *v
		fixed, e := fixSourceDir(ctx, src.Path, Module.Module.Path)
		if e != nil {
			logger.WithContext(ctx).Infof("invalid value of sources[%d].path: %v", i, e)
			continue

		}
		if fixed != src.Path {
			logger.WithContext(ctx).Debugf("Rewrite Source Path: %s => %s", v.Path, fixed)
		}
		src.Path = fixed
		sources = append(sources, &src)
	}
	Module.Sources = sources

	// Name default
	if Module.Name == "" {
		Module.Name = path.Base(Module.Module.Path)
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

// fixSourceDir attempts to fix given relative source directory and check if there is any source code in it
func fixSourceDir(ctx context.Context, dir string, module string) (string, error) {
	// remove ./, ../, etc
	dir = filepath.Clean(dir)
	// remove module prefix
	dir = strings.TrimPrefix(strings.ToLower(dir), strings.ToLower(module))
	// remove tailing /
	dir = strings.TrimSuffix(dir, "/")

	// find .go files
	srcs, e := cmdutils.LookupFiles("**/*.go", dir)
	if e != nil {
		return "", e
	} else if len(srcs) == 0 {
		return "", fmt.Errorf(`.go files not found in %s, ignoring`, dir)
	}

	return dir, nil
}
