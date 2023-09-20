package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

const (
	gOrderCleanup = GroupOrderProject + iota
)

/**********************
   Group
**********************/

type ProjectGroup struct {
	Option
}

func (g ProjectGroup) Order() int {
	return GroupOrderProject
}

func (g ProjectGroup) Name() string {
	return "Project"
}

func (g ProjectGroup) CustomizeTemplate() (TemplateOptions, error) {
	return nil, nil
}

func (g ProjectGroup) CustomizeData(data GenerationData) error {
	basic := ResolveEnabledLanaiModules(LanaiAppConfig, LanaiConsul, LanaiVault, LanaiRedis, LanaiTracing, LanaiDiscovery)
	pInit := data.ProjectMetadata()
	pInit.EnabledModules.Add(basic.Values()...)
	return nil
}

func (g ProjectGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	gOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&gOpt)
	}

	// Note: for backward compatibility, Default RegenMode is set to ignore
	gens := []Generator{
		newFileGenerator(gOpt, func(opt *FileOption) {
			opt.DefaultRegenMode = RegenModeIgnore
			opt.Prefix = "project."
		}),
		newDirectoryGenerator(gOpt, func(opt *DirOption) {
			opt.Patterns = []string{"cmd/**", "configs/**", "pkg/init/**"}
		}),
		newCleanupGenerator(gOpt, func(opt *CleanupOption) {
			opt.Order = gOrderCleanup
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}
