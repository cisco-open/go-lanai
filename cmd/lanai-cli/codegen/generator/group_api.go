package generator

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
    "fmt"
    "github.com/getkin/kin-openapi/openapi3"
)

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

    data, e := g.prepareData(&groupOpt)
    if e != nil {
        return nil, e
    }

    gens := []Generator{
        newDirectoryGenerator(func(opt *DirOption) {
            opt.Option = groupOpt
            opt.Data = data
        }),
        newApiGenerator(func(opt *ApiOption) {
            opt.Option = groupOpt
            opt.Data = data
        }),
        newApiGenerator(func(opt *ApiOption) {
            opt.Option = groupOpt
            opt.Data = data
            opt.Order = defaultApiStructOrder
            opt.Prefix = apiStructDefaultPrefix
        }),
        newFileGenerator(func(opt *FileOption) {
            opt.Option = groupOpt
            opt.Data = data
            opt.Prefix = "api-common."
        }),
        newApiVersionGenerator(func(opt *ApiVerOption) {
            opt.Option = groupOpt
            opt.Data = data
        }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}

func (g APIGroup) prepareData(opt *Option) (map[string]interface{}, error) {
    data := DataWithProject(&opt.Project)

    openAPIData, err := openapi3.NewLoader().LoadFromFile(opt.Components.Contract.Path)
    if err != nil {
        return nil, fmt.Errorf("error parsing OpenAPI file: %v", err)
    }
    data[KDataOpenAPI] = openAPIData
    return data, nil
}
