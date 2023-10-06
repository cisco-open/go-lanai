package generator

import (
    "io/fs"
)

// WithRegenRules Set re-generation rules, Fallback to default mode if no rules matches the output file
func WithRegenRules(rules RegenRules, defaultMode RegenMode) func(o *Option) {
    return func(option *Option) {
        option.RegenRules = rules
        if len(defaultMode) != 0 {
            option.DefaultRegenMode = defaultMode
        }
    }
}

func WithTemplateFS(templateFS fs.FS) func(o *Option) {
    return func(option *Option) {
        option.TemplateFS = templateFS
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



