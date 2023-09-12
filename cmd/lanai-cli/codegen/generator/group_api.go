package generator

import "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"

type APIGroup struct{}

func (g APIGroup) Order() int {
    return GroupOrderAPI
}

func (g APIGroup) Name() string {
    return "API"
}

func (g APIGroup) Generators(opts ...Options) ([]Generator, error) {
    groupOpt := DefaultOption
    for _, fn := range opts {
        fn(&groupOpt)
    }
    // TODO load OpenAPI and Data here instead of taking from options
    gens := []Generator{
        newDirectoryGenerator(func(opt *Option) { *opt = groupOpt }),
        newApiGenerator(func(opt *ApiGenOption) { opt.Option = groupOpt }),
        newApiGenerator(func(opt *ApiGenOption) {
            opt.Option = groupOpt
            opt.PriorityOrder = defaultApiStructOrder
            opt.Prefix = apiStructDefaultPrefix
        }),
        newFileGenerator(func(opt *FileGenOption) {
            opt.Option = groupOpt
            opt.Prefix = "api-common."
        }),
        newApiVersionGenerator(func(opt *Option) { *opt = groupOpt }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}
