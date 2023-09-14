package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"io/fs"
	"text/template"
)

const (
	GroupOrderAPI = iota
	GroupOrderProject
)

/***********************
	Generation Option
 ***********************/

var DefaultOption = Option{
	DefaultRegenMode: RegenModeIgnore,
}

type Options func(opt *Option)

type Option struct {
	// Project general project information
	Project Project

	// Components defines what to generate and their settings
	Components Components

	// TemplateFS filesystem containing templates. Could be embed.FS or os.DirFS
	TemplateFS fs.FS

	// DefaultRegenMode default output file operation mode during re-generation
	DefaultRegenMode RegenMode

	// RegenRules rules of output file operation mode during re-generation
	RegenRules RegenRules

	// internal fields, should not be set by external packages
	Template *template.Template
	//Data     map[string]interface{}
}

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
	// Data load and process group specific data. Such data might be useful for other group.
	// This function of all groups are called before any invocation of their Generators().
	// The results are merged into one map and served to Generators() for constructing generators.
	// Important: For each top-level component in Data should be owned by the group that generating it.
	// 			  Other groups can read it but should not change or replace it.
	Data(opts ...DataLoaderOptions) (map[string]interface{}, error)
	// Generators prepare generators with proper generator list based on options. Group can decide the group is not
	// applicable, in such case, Group can return empty list without error
	// The returned generators should be sorted.
	Generators(opts ...GenLoaderOptions) ([]Generator, error)
}

type DataLoaderOptions func(opt *DataLoaderOption)
type DataLoaderOption struct {
	Project    Project
	Components Components
}

type GenLoaderOptions func(opt *GenLoaderOption)
type GenLoaderOption struct {
	Option
	Template *template.Template
	Data     map[string]interface{}
}

/************************
	Generation Context
 ************************/

type GenerationContext struct {
	templatePath string
	filename     string
	regenMode    RegenMode
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
