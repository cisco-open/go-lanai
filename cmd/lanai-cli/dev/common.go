package dev

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
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