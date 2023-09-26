package generator

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/cmdutils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"io/fs"
	"path/filepath"
)

var logger = log.New("Codegen")
var globalCounter counter

func GenerateFiles(ctx context.Context, opts ...Options) error {
	generators := NewGenerators(opts...)
	return generators.Generate(ctx)
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
		OPAPolicyGroup{Option: ret.Option},
		SecurityGroup{Option: ret.Option},
		ProjectGroup{Option: ret.Option},
	}
	order.SortStable(ret.groups, order.UnorderedMiddleCompare)
	if ret.DefaultRegenMode != RegenModeIgnore {
		logger.Warnf(`Default Regen Mode is not "ignore". This is DANGEROUS!`)
	}
	return ret
}

func (g *Generators) Generate(ctx context.Context) error {
	// reset counter
	globalCounter = counter{}

	// load templates
	tmplOpts := make([]TemplateOptions, 0, len(g.groups))
	for _, group := range g.groups {
		switch opts, e := group.CustomizeTemplate(); {
		case e != nil:
			return e
		case opts != nil:
			tmplOpts = append(tmplOpts, opts)
		}
	}
	template, e := LoadTemplates(g.TemplateFS, tmplOpts...)
	if e != nil {
		return e
	}

	// populate data
	data := NewGenerationData(&g.Project)
	for _, group := range g.groups {
		if e := group.CustomizeData(data); e != nil {
			return e
		}
	}

	// prepare generators by groups
	generators := make([]Generator, 0, len(g.groups)*5)
	for _, group := range g.groups {
		gens, e := group.Generators(func(opt *GeneratorOption) {
			opt.Option = g.Option
			opt.Data = data
			opt.Template = template
		})
		if e != nil {
			return e
		}
		generators = append(generators, gens...)
	}

	// scan all templates
	tmpls := make([]TemplateDescriptor, 0, len(template.Templates()) * 2)
	e = fs.WalkDir(g.TemplateFS, ".", func(p string, d fs.DirEntry, err error) error {
		// load the files into the generator
		fi, e := d.Info()
		if e != nil {
			return e
		}
		tmpls = append(tmpls, TemplateDescriptor{Path: p, FileInfo: fi})
		return err
	})
	if e != nil {
		return e
	}

	// execute generators
	for _, gen := range generators {
		for _, tmpl := range tmpls {
			if err := gen.Generate(ctx, tmpl); err != nil {
				return err
			}
		}
	}

	// log summary
	for k, v := range globalCounter {
		if v == 0 {
			continue
		}
		path, e := filepath.Rel(cmdutils.GlobalArgs.OutputDir, k)
		if e != nil {
			path = k
		}
		count := fmt.Sprintf(`%d Files`, v)
		logger.Infof(`Generated %9s > %s`, count, path)
	}
	return nil
}

func NewGenerationData(p *Project) GenerationData {
	return map[string]interface{}{
		KDataProjectName: p.Name,
		KDataRepository:  p.Module,
		KDataProject: &ProjectMetadata{
			Project:        *p,
			EnabledModules: utils.NewStringSet(),
		},
		KDataLanaiModules: SupportedLanaiModules,
	}
}
