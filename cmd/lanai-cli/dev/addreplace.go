package dev

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"sort"
)

var (
	AddReplaceCmd = &cobra.Command{
		Use:                "deps-replace",
		Short:              "add the replace directive for a given module pattern and replace it with local path. Default to replace all local module",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunAddReplace,
	}
	AddReplaceArgs = AddReplaceArguments{
		Modules:     []string{"**"},
		SearchPaths: []string{"../"},
	}
)

type AddReplaceArguments struct {
	Modules     []string `flag:"modules,m" desc:"Comma delimited list of module pattern. e.g. cto-github.cisco.com/NFV-BU/**"`
	SearchPaths []string `flag:"paths,p" desc:"Comma delimited list of relative paths for searching local replacement"`
}

func init() {
	cmdutils.PersistentFlags(AddReplaceCmd, &AddReplaceArgs)
}

func RunAddReplace(cmd *cobra.Command, _ []string) error {
	cmdutils.ShCmdLogDisabled = true
	targetMod, e := cmdutils.GetGoMod(cmd.Context())
	if e != nil {
		return fmt.Errorf(`command need to run under a valid go module folder. cannot find "go.mod": %v`, e)
	}

	// validate pattern input
	toBeReplaced := utils.NewStringSet()
	for _, m := range AddReplaceArgs.Modules {
		if !doublestar.ValidatePathPattern(m) {
			return fmt.Errorf(`expected comma separated list of module pattern. e.g. cto-github.cisco.com/NFV-BU/**. But got "%s"`, m)
		}
		toBeReplaced.Add(m)
	}

	// find all available modules
	localModsMapping, e := resolveLocalMods(cmd.Context(), AddReplaceArgs.SearchPaths...)
	if e != nil {
		return e
	}
	localModsReversedMapping := map[string]string{}
	for k, v := range localModsMapping {
		localModsReversedMapping[v] = k
	}

	// find all required modules, including sub modules
	requires, e := resolveRequiredModules(cmd.Context())

	// resolve modules that need to be replaced
	toReplaceMods := utils.NewStringSet()
	for mod := range requires {
		if !pathMatches(mod, toBeReplaced) || mod == targetMod.Module.Path {
			continue
		}
		localPath, ok := localModsMapping[mod]
		if !ok {
			continue
		}
		toReplaceMods.Add(mod)

		// special treatment, if the local mod path is not git repo root, we need to add replace for its repo root to avoid error:
		// "ambiguous import: found package git/repo-root/sub-module in multiple modules:
		//  	git/repo-root/sub-module@version
		// 		git/repo-root@version"
		if cmdutils.IsGitRepoRoot(localPath) {
			continue
		}

		localRootMod := resolveGitRepoRootModFile(localPath)
		if rootMod, ok := localModsReversedMapping[localRootMod]; ok {
			toReplaceMods.Add(rootMod)
		} else {
			logger.WithContext(cmd.Context()).Warnf(`Unable to find Git repo root of %s: `, e)
		}
	}

	// add replace to target mod file
	var replaces []*cmdutils.Replace
	for reqPath := range toReplaceMods {
		relModPath := resolveLocalReplacePath(localModsMapping[reqPath], cmdutils.GlobalArgs.WorkingDir)
		replaces = append(replaces, &cmdutils.Replace{
			Old: cmdutils.Module{Path: reqPath},
			New: cmdutils.Module{Path: relModPath},
		})
	}
	sort.SliceStable(replaces, func(i, j int) bool {
		return replaces[i].Old.Path < replaces[j].Old.Path
	})
	for _, r := range replaces {
		logger.Debugf(`Replacing %s => %s`, r.Old.Path, r.New.Path)
	}

	cmdutils.ShCmdLogDisabled = false
	if e := cmdutils.SetReplace(cmd.Context(), replaces); e != nil {
		return fmt.Errorf(`unable to set replace: %v`, e)
	}

	// do "go mod tidy"
	if e := cmdutils.GoModTidy(cmd.Context(), nil); e != nil {
		return e
	}

	return nil
}

func resolveRequiredModules(ctx context.Context) (utils.StringSet, error) {
	subModFiles, e := findLocalGoModFiles(cmdutils.GlobalArgs.WorkingDir)
	if e != nil {
		return nil, fmt.Errorf(`command need to run under a valid go module folder. cannot find "go.mod": %v`, e)
	}
	requires := utils.NewStringSet()
	for _, modFile := range subModFiles {
		// using `go list -m`
		mods, e := cmdutils.FindModule(ctx, []cmdutils.GoCmdOptions{cmdutils.GoCmdModFile(modFile)}, "all")
		if e != nil {
			return nil, fmt.Errorf(`cannot open "go.mod": %v`, e)
		}
		for _, mod := range mods {
			requires.Add(mod.Path)
		}
	}
	return requires, nil
}

func resolveGitRepoRootModFile(subModPath string) string {
	repoRoot, e := cmdutils.FindGitRepoRoot(subModPath)
	if e != nil {
		return ""
	}
	repoRootModFile := filepath.Join(repoRoot, "go.mod")
	if _, e := os.Stat(repoRootModFile); e == nil {
		return repoRootModFile
	}
	return ""
}
