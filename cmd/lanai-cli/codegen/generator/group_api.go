package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
)

/**********************
   Data
**********************/

const (
	KDataOpenAPI = "OpenAPIData"
)

/**********************
   Group
**********************/

type APIGroup struct {
	Option
}

func (g APIGroup) Order() int {
	return GroupOrderAPI
}

func (g APIGroup) Name() string {
	return "API"
}

func (g APIGroup) CustomizeTemplate() (TemplateOptions, error) {
	return func(opt *TemplateOption) {
		// Note: API related functions are already added by default templates, we only need load it with configuration
		template_funcs.AddPredefinedRegexes(g.Components.Contract.Naming.RegExps)
	}, nil
}

func (g APIGroup) CustomizeData(data GenerationData) error {
	openAPIData, e := openapi3.NewLoader().LoadFromFile(g.Components.Contract.Path)
	if e != nil {
		return fmt.Errorf("error parsing OpenAPI file: %v", e)
	}
	data[KDataOpenAPI] = openAPIData

	pInit := data.ProjectMetadata()
	web := ResolveEnabledLanaiModules(LanaiWeb, LanaiActuator, LanaiSwagger)
	pInit.EnabledModules.Add(web.Values()...)
	return nil
}

func (g APIGroup) Generators(opts ...GeneratorOptions) ([]Generator, error) {
	genOpt := GeneratorOption{}
	for _, fn := range opts {
		fn(&genOpt)
	}

	gens := []Generator{
		newDirectoryGenerator(func(opt *DirOption) {
			opt.Option = g.Option
			opt.Data = genOpt.Data
			opt.Patterns = []string{"pkg/api/**", "pkg/controllers/**"}
		}),
		newApiGenerator(func(opt *ApiOption) {
			opt.Option = g.Option
			opt.Template = genOpt.Template
			opt.Data = genOpt.Data
		}),
		newApiGenerator(func(opt *ApiOption) {
			opt.Option = g.Option
			opt.Template = genOpt.Template
			opt.Data = genOpt.Data
			opt.Order = defaultApiStructOrder
			opt.Prefix = apiStructDefaultPrefix
		}),
		newFileGenerator(func(opt *FileOption) {
			opt.Option = g.Option
			opt.Template = genOpt.Template
			opt.Data = genOpt.Data
			opt.Prefix = "api-common."
		}),
		newApiVersionGenerator(func(opt *ApiVerOption) {
			opt.Option = g.Option
			opt.Template = genOpt.Template
			opt.Data = genOpt.Data
		}),
	}
	order.SortStable(gens, order.UnorderedMiddleCompare)
	return gens, nil
}
