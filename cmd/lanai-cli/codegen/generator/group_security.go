package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

/**********************
   Data
**********************/

const (
	KDataSecurity = "Security"
)

/**********************
   Group
**********************/

type SecurityGroup struct {
	Option
}

func (g SecurityGroup) Order() int {
	return GroupOrderSecurity
}

func (g SecurityGroup) Name() string {
	return "Security"
}

func (g SecurityGroup) CustomizeTemplate() (TemplateOptions, error) {
	return nil, nil
}

func (g SecurityGroup) CustomizeData(data GenerationData) error {
	if !g.isApplicable() {
		return nil
	}
	data[KDataSecurity] = g.Components.Security
	modules := make([]*LanaiModule, 0, 4)
	switch g.Components.Security.Authentication.Method {
	case AuthOAuth2:
		modules = append(modules, LanaiSecurity, LanaiResServer)
	}
	switch g.Components.Security.Access.Preset {
	case AccessPresetOPA:
		modules = append(modules, LanaiOPA)
	}
	sec := ResolveEnabledLanaiModules(modules...)
	data.ProjectMetadata().EnabledModules.Add(sec.Values()...)
	return nil
}

func (g SecurityGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
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
			opt.Prefix = "security."
		}),
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Patterns = []string{"configs/**", "pkg/init/**"}
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}

func (g SecurityGroup) isApplicable() bool {
	return len(g.Components.Security.Authentication.Method) != 0 && g.Components.Security.Authentication.Method != AuthNone
}
