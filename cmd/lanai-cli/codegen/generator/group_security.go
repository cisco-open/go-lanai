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
	pInit := data.ProjectMetadata()
	sec := ResolveEnabledLanaiModules(modules...)
	pInit.EnabledModules.Add(sec.Values()...)
	return nil
}

func (g SecurityGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	if !g.isApplicable() {
		return []Generator{}, nil
	}
	genOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&genOpt)
	}

	// Note: for backward compatibility, Default RegenMode is set to ignore
	gens := []Generator{
		newFileGenerator(func(opt *FileOption) {
			opt.Option = g.Option
			opt.Template = genOpt.Template
			opt.DefaultRegenMode = RegenModeIgnore
			opt.Data = genOpt.Data
			opt.Prefix = "security."
		}),
		newDirectoryGenerator(func(opt *DirOption) {
			opt.Option = g.Option
			opt.Data = genOpt.Data
			opt.Patterns = []string{"configs/**", "pkg/init/**"}
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}

func (g SecurityGroup) isApplicable() bool {
	return len(g.Components.Security.Authentication.Method) != 0 && g.Components.Security.Authentication.Method != AuthNone
}
