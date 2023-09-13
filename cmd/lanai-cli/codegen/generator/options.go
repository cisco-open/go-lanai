package generator

import (
    "io/fs"
    "text/template"
)

var DefaultOption = Option{}

type Options func(opt *Option)

type Option struct {
    Template *template.Template
    // Data used
    // Deprecated: Data should be calculated from Project, Components, etc.
    Data map[string]interface{}

    // Project general project information
    Project Project

    // Components defines what to generate and their settings
    Components Components

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

    // DefaultRegenMode default output file operation mode during re-generation
    DefaultRegenMode RegenMode

    // RegenRules rules of output file operation mode during re-generation
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

// WithData
// Deprecated, use WithConfig instead
func WithData(data map[string]interface{}) func(o *Option) {
    return func(o *Option) {
        o.Data = data
    }
}

// WithProject general information about the project to generate
func WithProject(project Project) func(o *Option) {
    return func(o *Option) {
        o.Project = project
    }
}

// WithComponents defines what to generate and their settings
func WithComponents(comps Components) func(o *Option) {
    return func(o *Option) {
        o.Components = comps
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

