package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"io/fs"
	"text/template"
)

var logger = log.New("Codegen.generator")

type Generator interface {
	Generate(tmplPath string, dirEntry fs.DirEntry) error
}

type Generators struct {
	generators []Generator
}

type Option struct {
	Template *template.Template
	Data     map[string]interface{}
	FS       fs.FS
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

func NewGenerators(opts ...func(*Option)) Generators {
	return Generators{
		generators: []Generator{
			newApiGenerator(opts...),
			newProjectGenerator(opts...),
			newDirectoryGenerator(opts...),
		},
	}
}

func (g *Generators) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	for _, gen := range g.generators {
		if err := gen.Generate(tmplPath, dirEntry); err != nil {
			return err
		}
	}

	return nil
}
