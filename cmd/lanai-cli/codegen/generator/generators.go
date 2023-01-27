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
}

type Rules struct {
	Regeneration string `yaml:"regeneration"`
}

type templateInfo struct {
	tmplPath string
	dirEntry fs.DirEntry
}

type Option struct {
	Template      *template.Template
	Data          map[string]interface{}
	FS            fs.FS
	PriorityOrder int
	Prefix        string
	RegenRule     string
	Rules         map[string]Rules
}

func WithRules(rules map[string]Rules) func(o *Option) {
	return func(option *Option) {
		option.Rules = rules
	}
}
func WithRegenerationRule(globalRegenerationRule string) func(o *Option) {
	return func(option *Option) {
		option.RegenRule = globalRegenerationRule
	}
}
func WithFS(filesystem fs.FS) func(o *Option) {
	return func(option *Option) {
		option.FS = filesystem
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

func WithPriorityOrder(order int) func(o *Option) {
	return func(o *Option) {
		o.PriorityOrder = order
	}
}

func WithPrefix(prefix string) func(o *Option) {
	return func(o *Option) {
		o.Prefix = prefix
	}
}
func NewGenerators(opts ...func(*Option)) Generators {
	ret := Generators{
		generators: []Generator{
			newApiGenerator(append(opts, WithPrefix("api-struct."), WithPriorityOrder(defaultApiStructOrder))...),
			newApiGenerator(opts...),
			newProjectGenerator(opts...),
			newDirectoryGenerator(opts...),
			newVersionGenerator(opts...),
			newDeleteGenerator(opts...),
		},
	}
	order.SortStable(ret.generators, order.OrderedLastCompare)

	return ret
}

func (g *Generators) Generate() error {
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
