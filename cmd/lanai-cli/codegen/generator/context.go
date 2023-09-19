package generator

import (
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
    "cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"io/fs"
	"text/template"
)

/***************************
	Global Vars/Constants
 ***************************/

// Global Keys in template's context data as map
const (
	KDataProjectName  = "ProjectName"
	KDataRepository   = "Repository"
	KDataProject      = "Project"
	KDataProjectInit  = "ProjectInit"
	KDataLanaiModules = "LanaiModules"
)

const (
	GroupOrderAPI = iota
	GroupOrderOPAPolicy
	GroupOrderSecurity
	GroupOrderProject
)

/***********************
	Generation Option
 ***********************/

var DefaultOption = Option{
	DefaultRegenMode: RegenModeIgnore,
	Components: Components{
		Contract: Contract{},
		Security: Security{
			Authentication: Authentication{
				Method: AuthOAuth2,
			},
			Access:         Access{
				Preset: AccessPresetFreestyle,
			},
		},
	},
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

	// CustomizeTemplate is used to customize *template.Template for group specific need
	CustomizeTemplate() (TemplateOptions, error)

	// CustomizeData customize data and load group specific data. All data is shared with other groups.
	// This function of all groups are called in order before any invocation of their Generators().
	// The results are propagate to other groups and served to Generators() for constructing generators.
	// Important: For each top-level component in Data should be owned by the group that generating it.
	// 			  Other groups can read it but should not change or replace it.
	//			  The exception is "ProjectInit" object, which all groups can add to it
	CustomizeData(data GenerationData) error
	// Generators prepare generators with proper generator list based on options. Group can decide the group is not
	// applicable, in such case, Group can return empty list without error
	// The returned generators should be sorted.
	Generators(opts ...GeneratorOptions) ([]Generator, error)
}

//goland:noinspection GoNameStartsWithPackageName
type GeneratorOptions func(opt *GeneratorOption)

//goland:noinspection GoNameStartsWithPackageName
type GeneratorOption struct {
	Template *template.Template
	Data     map[string]interface{}
}

/************************
	Generation Context
 ************************/

type GenerationData map[string]interface{}

func (d GenerationData) ProjectMetadata() *ProjectInit {
    _, ok := d[KDataProjectInit]
    if !ok {
        d[KDataProjectInit] = &ProjectInit{
            EnabledModules: utils.NewStringSet(),
        }
    }
    return d[KDataProjectInit].(*ProjectInit)
}

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
