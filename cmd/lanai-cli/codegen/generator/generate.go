package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"io/fs"
)

var logger = log.New("Codegen.generator")

const (
	defaultProjectPriorityOrder = iota
	defaultApiStructOrder
	defaultApiPriorityOrder
	defaultDeletePriorityOrder
)

func GenerateFiles(opts ...Options) error {
	generators := NewGenerators(opts...)
	return generators.Generate()
}

type Generators struct {
	Option
	groups []Group
}

func NewGenerators(opts ...Options) Generators {
	ret := Generators{
		Option: DefaultOption,
	}
	for _, fn := range opts {
		fn(&ret.Option)
	}
	ret.groups = []Group{
		APIGroup{Option: ret.Option},
		SecurityGroup{Option: ret.Option},
		ProjectGroup{Option: ret.Option},
	}
	order.SortStable(ret.groups, order.UnorderedMiddleCompare)
	if ret.DefaultRegenMode != RegenModeIgnore {
		logger.Warnf(`Default Regen Mode is not "ignore". This is DANGEROUS!`)
	}
	return ret
}

func (g *Generators) Generate() error {

	// load templates
	tmplOpts := make([]TemplateOptions, 0, len(g.groups))
	for _, group := range g.groups {
		opts, e := group.CustomizeTemplate()
		if e != nil {
			return e
		}
		tmplOpts = append(tmplOpts, opts)
	}
	template, e := LoadTemplates(g.TemplateFS, tmplOpts...)
	if e != nil {
		return e
	}

	// populate data
	data := newGenerationData(&g.Project)
	for _, group := range g.groups {
		if e := group.CustomizeData(data); e != nil {
			return e
		}
	}

	// prepare generators by groups
	generators := make([]Generator, 0, len(g.groups)*5)
	for _, group := range g.groups {
		gens, e := group.Generators(func(opt *GeneratorOption) {
			opt.Data = data
			opt.Template = template
		})
		if e != nil {
			return e
		}
		generators = append(generators, gens...)
	}

	// scan all templates
	tmpls := make(map[string]fs.FileInfo)
	e = fs.WalkDir(g.TemplateFS, ".", func(p string, d fs.DirEntry, err error) error {
		// load the files into the generator
		fi, e := d.Info()
		if e != nil {
			return e
		}
		tmpls[p] = fi
		return err
	})
	if e != nil {
		return e
	}

	// execute generators
	for _, gen := range generators {
		for path, info := range tmpls {
			if err := gen.Generate(path, info); err != nil {
				return err
			}
		}
	}
	return nil
}

func newGenerationData(p *Project) GenerationData {
	return map[string]interface{}{
		KDataProjectName:  p.Name,
		KDataRepository:   p.Module,
		KDataProject:      p,
		KDataLanaiModules: SupportedLanaiModules,
		KDataProjectInit: &ProjectInit{
			EnabledModules: utils.NewStringSet(),
		},
	}
}
