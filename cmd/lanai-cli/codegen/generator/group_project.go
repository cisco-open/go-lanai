package generator

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"

type ProjectGroup struct {}

func (g ProjectGroup) Order() int {
    return GroupOrderProject
}

func (g ProjectGroup) Name() string {
    return "API"
}

func (g ProjectGroup) Generators(opts ...Options) ([]Generator, error) {
    groupOpt := DefaultOption
    for _, fn := range opts {
        fn(&groupOpt)
    }
    // TODO load Project scaffolding Data here instead of taking from options
    gens := []Generator{
        newFileGenerator(func(opt *FileGenOption) { opt.Option = groupOpt }),
        newDirectoryGenerator(func(opt *Option) { *opt = groupOpt }),
        newDeleteGenerator(func(opt *DeleteOption) { opt.Option = groupOpt }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}
