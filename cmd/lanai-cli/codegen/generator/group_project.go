package generator

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

type ProjectGroup struct {}

func (g ProjectGroup) Order() int {
    return GroupOrderProject
}

func (g ProjectGroup) Name() string {
    return "API"
}

func (g ProjectGroup) Data(opts ...DataLoaderOptions) (map[string]interface{}, error) {
    // TODO load Project scaffolding Data here instead of taking from options
    return map[string]interface{}{}, nil
}

func (g ProjectGroup) Generators(opts ...GenLoaderOptions) ([]Generator, error) {
    loaderOpt := GenLoaderOption{
        Option: DefaultOption,
    }
    for _, fn := range opts {
        fn(&loaderOpt)
    }

    // Note: for backward compatibility, Default RegenMode is set to ignore
    gens := []Generator{
        newFileGenerator(func(opt *FileOption) {
            opt.Option = loaderOpt.Option
            opt.DefaultRegenMode = RegenModeIgnore
            opt.Data = loaderOpt.Data
        }),
        newDirectoryGenerator(func(opt *DirOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
            opt.Patterns = []string{"cmd/**", "configs/**", "pkg/init/**"}
        }),
        newDeleteGenerator(func(opt *DeleteOption) { opt.Option = loaderOpt.Option }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}

