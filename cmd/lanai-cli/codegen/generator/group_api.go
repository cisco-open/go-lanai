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

func (g APIGroup) Data(opts ...DataLoaderOptions) (map[string]interface{}, error) {
    opt := DataLoaderOption{}
    for _, fn := range opts {
        fn(&opt)
    }

    openAPIData, err := openapi3.NewLoader().LoadFromFile(opt.Components.Contract.Path)
    if err != nil {
        return nil, fmt.Errorf("error parsing OpenAPI file: %v", err)
    }

    return map[string]interface{}{
        KDataOpenAPI: openAPIData,
    }, nil
}

func (g APIGroup) Generators(opts ...GenLoaderOptions) ([]Generator, error) {
    loaderOpt := GenLoaderOption{
        Option: DefaultOption,
    }
    for _, fn := range opts {
        fn(&loaderOpt)
    }

    gens := []Generator{
        newDirectoryGenerator(func(opt *DirOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
            opt.Patterns = []string{"pkg/api/**", "pkg/controllers/**"}
        }),
        newApiGenerator(func(opt *ApiOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
        }),
        newApiGenerator(func(opt *ApiOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
            opt.Order = defaultApiStructOrder
            opt.Prefix = apiStructDefaultPrefix
        }),
        newFileGenerator(func(opt *FileOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
            opt.Prefix = "api-common."
        }),
        newApiVersionGenerator(func(opt *ApiVerOption) {
            opt.Option = loaderOpt.Option
            opt.Data = loaderOpt.Data
        }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}
