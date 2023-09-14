package generator

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
)

type ProjectGroup struct{
    Option
}

func (g ProjectGroup) Order() int {
    return GroupOrderProject
}

func (g ProjectGroup) Name() string {
    return "API"
}

func (g ProjectGroup) CustomizeTemplate() (TemplateOptions, error) {
    return func(opt *TemplateOption) {}, nil
}

func (g ProjectGroup) Data() (map[string]interface{}, error) {
    // TODO load Project scaffolding Data here instead of taking from options
    return map[string]interface{}{}, nil
}

func (g ProjectGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
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
        }),
        newDirectoryGenerator(func(opt *DirOption) {
            opt.Option = g.Option
            opt.Data = genOpt.Data
            opt.Patterns = []string{"cmd/**", "configs/**", "pkg/init/**"}
        }),
        newDeleteGenerator(func(opt *DeleteOption) { opt.Option = g.Option }),
    }
    order.SortStable(gens, order.UnorderedMiddleCompare)
    return gens, nil
}

