package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"io/fs"
	"text/template"
)

var logger = log.New("Codegen.generator")

const (
	defaultProjectPriorityOrder = iota
	defaultApiStructOrder
	defaultApiPriorityOrder
	defaultDeletePriorityOrder
)

type Generator interface {
	Generate(tmplPath string, dirEntry fs.DirEntry) error
}

type Generators struct {
	generators  []Generator
	loadedPaths []templateInfo
	templateFS  fs.FS
}

type templateInfo struct {
	tmplPath string
	dirEntry fs.DirEntry
}

type Option struct {
	Template *template.Template
	// Data used
	Data map[string]interface{}
	// FS
	// Deprecated: This is incorrect and confusing: We need two FSs, one is template FS as input (could be embedded or OS dir),
	// and the other one is output FS.
	// Use OutputFS and TemplateFS
	FS fs.FS
	// OutputFS filesystem for output files. Generators assume the filesystem's root is the project root
	// TODO: This value is currently not used by generators. Need to update generators to support this
	OutputFS fs.FS
	// TemplateFS filesystem containing templates. Could be embed.FS or os.DirFS
	TemplateFS fs.FS

	// PriorityOrder When applicable, indicate the execution order of each generator
	// Deprecated: similar to Prefix, this value is not applicable to all generators. When it's applicable, it would be
	// different per generator.
	// When applicable, use generator's own options
	PriorityOrder int

	// Prefix prefix of template file that individual Generator should pick up. e.g. FileGenerator would
	// pick up any template with "project.*.tmpl"
	// Deprecated: This is incorrect and confusing: If all generators uses same Option, this field is useless,
	// 	  because all generators would either ignore this value or requires different value.
	// Solution:
	//	1. Generators should have their own "Option" struct, embedding this struct. "Prefix" should be defined in their own
	//     "Option" struct if "Prefix" is applicable to that particular generator
	// 	2. Change the name to "TemplatePrefix" to avoid confusion
	Prefix           string
	DefaultRegenMode RegenMode
	RegenRules       RegenRules
}

// WithRegenRules Set re-generation rules, Fallback to default mode if no rules matches the output file
func WithRegenRules(rules RegenRules, defaultMode RegenMode) func(o *Option) {
	return func(option *Option) {
		option.RegenRules = rules
		if len(defaultMode) != 0 {
			option.DefaultRegenMode = defaultMode
		}
	}
}

// WithFS
// Deprecated: use WithTemplateFS and WithOutputFS instead
func WithFS(filesystem fs.FS) func(o *Option) {
	return func(option *Option) {
		option.FS = filesystem
	}
}

// WithOutputFS set output filesystem
// TODO: This value is currently not used by generators. Need to update generators to support this
func WithOutputFS(outputFS fs.FS) func(o *Option) {
	return func(option *Option) {
		option.OutputFS = outputFS
	}
}

func WithTemplateFS(templateFS fs.FS) func(o *Option) {
	return func(option *Option) {
		option.TemplateFS = templateFS
	}
}

func WithData(data map[string]interface{}) func(o *Option) {
	return func(o *Option) {
		o.Data = data
	}
}

func WithTemplate(template *template.Template) func(o *Option) {
	return func(o *Option) {
		o.Template = template
	}
}

// WithPriorityOrder
// Deprecated: use generator's own option to set it, if applicable
func WithPriorityOrder(order int) func(o *Option) {
	return func(o *Option) {
		o.PriorityOrder = order
	}
}

// WithPrefix
// Deprecated: Prefix doesn't apply to all generators
func WithPrefix(prefix string) func(o *Option) {
	return func(o *Option) {
		o.Prefix = prefix
	}
}
func NewGenerators(opts ...func(*Option)) Generators {
	genOpt := Option{}
	for _, fn := range opts {
		fn(&genOpt)
	}

	apiGenerators := []Generator{
		newApiGenerator(func(opt *ApiGenOption) {
			opt.Option = genOpt
			opt.Prefix = apiStructDefaultPrefix
			opt.PriorityOrder = defaultApiStructOrder
		}),
		newApiGenerator(func(opt *ApiGenOption) { opt.Option = genOpt }),
		newFileGenerator(func(opt *FileGenOption) {
			opt.Option = genOpt
			opt.Prefix = "api-common."
		}),
		newApiVersionGenerator(opts...),
	}

	skeletonGenerators := []Generator {
		newFileGenerator(func(opt *FileGenOption) { opt.Option = genOpt }),
		newDirectoryGenerator(opts...),
		newDeleteGenerator(func(opt *DeleteOption) { opt.Option = genOpt }),
	}

	return Generators{
		generators: mergeGenerators(apiGenerators, skeletonGenerators),
		templateFS: genOpt.TemplateFS,
	}
}

func mergeGenerators(generators ...[]Generator) []Generator {
	length := 0
	for i := range generators {
		length += len(generators[i])
	}
	gens := make([]Generator, 0, length)
	for _, slice := range generators {
		gens = append(gens, slice...)
	}
	order.SortStable(gens, order.OrderedLastCompare)
	return gens
}

func (g *Generators) Generate() error {
	// scan all templates
	e := fs.WalkDir(g.templateFS, ".",
		func(p string, d fs.DirEntry, err error) error {
			// load the files into the generator
			g.Load(p, d)
			return nil
		})
	if e != nil {
		return e
	}

	// execute generators
	for _, gen := range g.generators {
		for _, loadedPath := range g.loadedPaths {
			if err := gen.Generate(loadedPath.tmplPath, loadedPath.dirEntry); err != nil {
				return err
			}
		}
	}

	return nil
}

func (g *Generators) Load(tmplPath string, dirEntry fs.DirEntry) {
	g.loadedPaths = append(g.loadedPaths, templateInfo{
		tmplPath: tmplPath,
		dirEntry: dirEntry,
	})
}
