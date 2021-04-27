package deps

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"errors"
	"fmt"
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
)

func RunUpdateDep(cmd *cobra.Command, _ []string) error {
	//process input args to see which module's dependency needs to be updated
	moduleToBranch := make(map[string]string)
	for _, arg := range Args.Modules {
		mbr := strings.Split(arg, "@")
		if len(mbr) != 2 {
			logger.Errorf("%s doesn't follow the module@branch format", mbr)
			return errors.New("can't parse module path")
		}
		m := mbr[0]
		br := mbr[1]

		moduleToBranch[m] = br
	}

	dropped, e := cmdutils.DropInvalidReplace(cmd.Context())
	if e != nil {
		return fmt.Errorf("unable to temporarily drop invalid 'replace' in go.mod: %v", e)
	}

	// update their dependencies
	for module, branch := range moduleToBranch {
		logger.Infof("processing %s@%s", module, branch)
		e := cmdutils.GoGet(cmd.Context(), module, branch)
		if e != nil {
			return nil
		}
	}

	// go mod tidy to update implicit dependencies changes
	if e := cmdutils.GoModTidy(cmd.Context()); e != nil {
		return e
	}

	if e := cmdutils.RestoreInvalidReplace(cmd.Context(), dropped); e != nil {
		return fmt.Errorf("unable to restore temporarily dropped 'replace' in go.mod: %v", e)
	}

	// mark changes if requested
	msg := fmt.Sprintf("updated versions of private modules")
	tag, e := markChangesIfRequired(cmd.Context(), msg, cmdutils.GitFilePattern("go.mod", "go.sum"))
	if e == nil && tag != "" {
		logger.WithContext(cmd.Context()).Infof(`Modified go.mod/go.sum are tagged with Git Tag [%s]`, tag)
	}
	return e
}