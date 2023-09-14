package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
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
	groups     []Group
}

func NewGenerators(opts ...Options) Generators {
	ret := Generators{
		Option: DefaultOption,
		groups: []Group{
			APIGroup{},
			ProjectGroup{},
		},
	}
	for _, fn := range opts {
		fn(&ret.Option)
	}
	order.SortStable(ret.groups, order.UnorderedMiddleCompare)
	if ret.DefaultRegenMode != RegenModeIgnore {
		logger.Warnf(`Default Regen Mode is not "ignore". This is DANGEROUS!`)
	}
	return ret
}

func (g *Generators) Generate() error {

	// populate data
	data := newCommonData(&g.Project)
	for _, group := range g.groups {
		groupData, e := group.Data(func(opt *DataLoaderOption) {
			opt.Project = g.Project
			opt.Components = g.Components
		})
		if e != nil {
			return e
		}
		g.shallowMerge(data, groupData)
	}

	// prepare generators by groups
	generators := make([]Generator, 0, len(g.groups) * 5)
	for _, group := range g.groups {
		gens, e := group.Generators(func(opt *GenLoaderOption) {
			opt.Option = g.Option
			opt.Data = data
			opt.Template = g.Template
		})
		if e != nil {
			return e
		}
		generators = append(generators, gens...)
	}

	// scan all templates
	tmpls := make(map[string]fs.FileInfo)
	e := fs.WalkDir(g.TemplateFS, ".", func(p string, d fs.DirEntry, err error) error {
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

func (g *Generators) shallowMerge(dest, src map[string]interface{}) {
	for k, v := range src {
		dest[k] = v
	}
}