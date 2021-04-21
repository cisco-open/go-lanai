package deps

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"errors"
	"github.com/spf13/cobra"
	"strings"
)

var (
	DropReplaceCmd = &cobra.Command{
		Use:                "drop-replace",
		Short:              "drop the replace directive for a given module and update go.sum",
		FParseErrWhitelist: cobra.FParseErrWhitelist{UnknownFlags: true},
		RunE:               RunDropReplace,
	}
	DropReplaceArgs = DropReplaceArguments{
		Modules: []string{},
	}
)

type DropReplaceArguments struct {
	Modules  []string `flag:"modules,m" desc:"Comma delimited list of modules"`
}

func init() {
	cmdutils.PersistentFlags(DropReplaceCmd, &DropReplaceArgs)
}

func RunDropReplace(cmd *cobra.Command, _ []string) error {
	toBeReplaced := utils.NewStringSet()

	for _, arg := range DropReplaceArgs.Modules {
		m := strings.Split(arg, "@") //because arg can be module or module:version. if it's the latter, we take the part before :
		if len(m) != 1 && len(m) != 2 {
			return errors.New("input should be a comma separated list of module or module@version strings")
		}
		toBeReplaced.Add(m[0])
	}

	mod, err := cmdutils.GetGoMod(cmd.Context())
	if err != nil {
		return err
	}

	changed := false
	for _, replace := range mod.Replace {
		logger.Infof("found replace for %s, %s", replace.Old.Path, replace.Old.Version)

		//we only drop the replace for the module whose dependency we updated
		//there may be replace that are pointing to other module (not local)
		if toBeReplaced.Has(replace.Old.Path) {
			err = cmdutils.DropReplace(cmd.Context(), replace.Old.Path, replace.Old.Version)
			if err != nil {
				return err
			}
			changed = true
		}
	}

	if changed {
		if e := cmdutils.GoModTidy(cmd.Context()); e != nil {
			return e
		}
	}
	return err
}
