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
	return ret
}

func (g *Generators) Generate() error {

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

	// execute generators by groups
	for _, group := range g.groups {
		gens, e := group.Generators(func(opt *Option) { *opt = g.Option })
		if e != nil {
			return e
		}
		for _, gen := range gens {
			for path, info := range tmpls {
				if err := gen.Generate(path, info); err != nil {
					return err
				}
			}
		}
	}
	return nil
}
