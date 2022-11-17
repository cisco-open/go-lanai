package dev

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/spf13/cobra"
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
	localMods, e := resolveLocalMods(cmd.Context(), AddReplaceArgs.SearchPaths...)
	if e != nil {
		return e
	}

	// find all required modules (using `go list -m`), including sub modules
	subModFiles, e := findLocalGoMods(cmdutils.GlobalArgs.WorkingDir)
	if e != nil {
		return fmt.Errorf(`command need to run under a valid go module folder. cannot find "go.mod": %v`, e)
	}
	requires := utils.NewStringSet()
	for _, modFile := range subModFiles {
		mods, e := cmdutils.FindModule(cmd.Context(), []cmdutils.GoCmdOptions{cmdutils.GoCmdModFile(modFile)}, "all")
		if e != nil {
			return fmt.Errorf(`cannot open "go.mod": %v`, e)
		}
		for _, mod := range mods {
			requires.Add(mod.Path)
		}
	}

	// add replace to target mod file
	var replaces []*cmdutils.Replace
	for reqPath := range requires {
		if !pathMatches(reqPath, toBeReplaced) || reqPath == targetMod.Module.Path {
			continue
		}
		modPath, ok := localMods[reqPath]
		if !ok {
			continue
		}
		relModPath := resolveLocalReplacePath(modPath, cmdutils.GlobalArgs.WorkingDir)
		logger.Debugf(`Replacing %s => %s`, reqPath, relModPath)
		replaces = append(replaces, &cmdutils.Replace{
			Old: cmdutils.Module{Path: reqPath},
			New: cmdutils.Module{Path: relModPath},
		})
	}

	cmdutils.ShCmdLogDisabled = false
	if e := cmdutils.SetReplace(cmd.Context(), replaces); e != nil {
		return fmt.Errorf(`unable to set replace: %v`, e)
	}

	// do "go mod tidy"
	if e := cmdutils.GoModTidy(cmd.Context()); e != nil {
		return e
	}

	return nil
}
