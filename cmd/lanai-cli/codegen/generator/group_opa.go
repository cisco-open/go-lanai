package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"strings"
)

/**********************
   Data
**********************/

const (
	KDataOPAPolicy = "OPAPolicy"
)

type OPAPolicyData struct {
	APIPackage string
}

/**********************
   Group
**********************/

// OPAPolicyGroup generate placeholder for OPA policies.
// This group is not responsible to setup security init source code. SecurityGroup would do that
type OPAPolicyGroup struct {
	Option
}

func (g OPAPolicyGroup) Order() int {
	return GroupOrderOPAPolicy
}

func (g OPAPolicyGroup) Name() string {
	return "OPA Policy"
}

func (g OPAPolicyGroup) CustomizeTemplate() (TemplateOptions, error) {
	return nil, nil
}

func (g OPAPolicyGroup) CustomizeData(data GenerationData) error {
	if !g.isApplicable() {
		return nil
	}
	data[KDataOPAPolicy] = OPAPolicyData{
		APIPackage: strings.ReplaceAll(g.Project.Name, "-", "_") + "_api",
	}
	return nil
}

func (g OPAPolicyGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	if !g.isApplicable() {
		return []Generator{}, nil
	}

	gOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&gOpt)
	}

	// Note: for backward compatibility, Default RegenMode is set to ignore
	gens := []Generator{
		newFileGenerator(gOpt, func(opt *FileOption) {
			opt.DefaultRegenMode = RegenModeIgnore
			opt.Prefix = "opa"
		}),
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Matcher = isDir().And(matchPatterns("policies/**"))
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}

func (g OPAPolicyGroup) isApplicable() bool {
	return g.Components.Security.Access.Preset == AccessPresetOPA
}
