package generator

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
    "fmt"
    "github.com/getkin/kin-openapi/openapi3"
)

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
    data, e := g.prepareData(&groupOpt)
    if e != nil {
        return nil, e
    }

    gens := []Generator{
        newFileGenerator(func(opt *FileOption) {
            opt.Option = groupOpt
            opt.Data = data
        }),
        newDirectoryGenerator(func(opt *DirOption) {
            opt.Option = groupOpt
            opt.Data = data
        }),
        newDeleteGenerator(func(opt *DeleteOption) { opt.Option = groupOpt }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}

func (g ProjectGroup) prepareData(opt *Option) (map[string]interface{}, error) {
    data := DataWithProject(&opt.Project)

    openAPIData, err := openapi3.NewLoader().LoadFromFile(opt.Components.Contract.Path)
    if err != nil {
        return nil, fmt.Errorf("error parsing OpenAPI file: %v", err)
    }
    data[KDataOpenAPI] = openAPIData
    return data, nil
}
