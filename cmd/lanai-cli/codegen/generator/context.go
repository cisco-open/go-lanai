package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"io/fs"
)

const (
	GroupOrderAPI = iota
	GroupOrderProject
)

/*******************
	Interfaces
 *******************/

// Generator interface for various code generation
type Generator interface {
	// Generate files based on given template.
	// Provided template could be
	// 	- Path to a loaded template file
	// 	- Directory path in the template FS
	//
	// Generator's implementation decide what to with the given template, including:
	// 	- Creating directories
	//  - Generating single file
	//  - Generating multiple files
	Generate(tmplPath string, tmplInfo fs.FileInfo) error
}

// Group is a collection of concrete generator instances that work together to generate set of files and directories
// for one particular components.
// Group should be always named and ordered
type Group interface {
	order.Ordered

	// Name unique identifier of the group
	Name() string
	// Generators prepare generators with proper generation data based on options. Group can decide the group is not
	// applicable, in such case, Group can return empty list without error
	// The returned generators should be sorted.
	Generators(opts ...Options) ([]Generator, error)
}

/************************
	Generation Context
 ************************/

type GenerationContext struct {
	templatePath string
	filename  string
	regenMode RegenMode
	//	 Add the template (template.Template) here
	model interface{}
}

func NewGenerationContext(templatePath string, filename string, regenMode RegenMode, model interface{}) *GenerationContext {
	return &GenerationContext{
		templatePath: templatePath,
		filename:     filename,
		regenMode:    regenMode,
		model:        model,
	}
}
