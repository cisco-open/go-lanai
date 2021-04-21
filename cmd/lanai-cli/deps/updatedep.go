package deps

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

var (
	UpdateDepCmd = &cobra.Command{
		Use:                "update",
		Short:              "Update module dependencies with given branches and update go.sum",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunUpdateDep,
	}
	UpdateDepArgs = UpdateDepArguments{
		ModuleBranches: []string{},
	}
)

type UpdateDepArguments struct {
	ModuleBranches  []string `flag:"modules,m" desc:"Comma delimited list of module@branch"`
}

func init() {
	cmdutils.PersistentFlags(UpdateDepCmd, &UpdateDepArgs)
}

func RunUpdateDep(cmd *cobra.Command, _ []string) error {
	//process input args to see which module's dependency needs to be updated
	moduleToBranch := make(map[string]string)
	modules := []string{}
	for _, arg := range UpdateDepArgs.ModuleBranches {
		mbr := strings.Split(arg, "@")
		if len(mbr) != 2 {
			logger.Errorf("%s doesn't follow the module:path format", mbr)
			return errors.New("can't parse module path")
		}
		m := mbr[0]
		br := mbr[1]
		logger.Infof("processing %s@%s", m, br)

		moduleToBranch[m] = br
		modules = append(modules, m)
	}

	// update their dependencies
	for module, branch := range moduleToBranch {
		err := cmdutils.GoGet(cmd.Context(), module, branch)
		if err != nil {
			return nil
		}
	}

	// go mod tidy to update implicit dependencies changes
	if e := cmdutils.GoModTidy(cmd.Context()); e != nil {
		return e
	}
	return nil
}