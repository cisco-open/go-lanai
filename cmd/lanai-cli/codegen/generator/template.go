package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/template_funcs"
	"github.com/bmatcuk/doublestar/v4"
	"io/fs"
	"text/template"
)

type TemplateFuncProvider func(funcMap template.FuncMap) (template.FuncMap, error)
type TemplatePreHook func(tmpl *template.Template, tmplPaths []string) (*template.Template, []string, error)
type TemplatePostHook func(tmpl *template.Template) (*template.Template, error)

type TemplateOptions func(opt *TemplateOption)
type TemplateOption struct {
	Pattern       string
	FuncProviders []TemplateFuncProvider
	PreloadHooks  []TemplatePreHook
	PostHooks     []TemplatePostHook
}

func LoadTemplates(fsys fs.FS, opts ...TemplateOptions) (tmpl *template.Template, err error) {
	opt := TemplateOption{
		Pattern: "**/*.tmpl",
		FuncProviders: []TemplateFuncProvider{defaultTemplateFuncProvider()},
	}
	for _, fn := range opts {
		fn(&opt)
	}

	// configure functions
	fnMap := template.FuncMap{}
	for _, provider := range opt.FuncProviders {
		if fnMap, err = provider(fnMap); err != nil {
			return
		}
	}

	// search for templates
	filenames, e := findTemplateFiles(fsys, opt.Pattern)
	if e != nil {
		return nil, e
	}

	// create template and prepare for loading
	tmpl = template.New("templates").Funcs(fnMap)
	for _, hookFn := range opt.PreloadHooks {
		if tmpl, filenames, err = hookFn(tmpl, filenames); err != nil {
			return nil, err
		}
	}

	// load templates
	for _, path := range filenames {
		content, e := fs.ReadFile(fsys, path)
		if e != nil {
			return nil, e
		}
		if tmpl, err = tmpl.New(path).Parse(string(content)); err != nil {
			return nil, err
		}
	}

	// post process
	for _, hookFn := range opt.PostHooks {
		if tmpl, err = hookFn(tmpl); err != nil {
			return nil, err
		}
	}
	return tmpl, nil
}

func defaultTemplateFuncProvider() TemplateFuncProvider {
	return func(funcMap template.FuncMap) (template.FuncMap, error) {
		template_funcs.Load()
		//template_funcs.AddPredefinedRegexes(loaderOptions.InitialRegexes)
		for _, fm := range template_funcs.TemplateFuncMaps {
			for k, v := range fm {
				funcMap[k] = v
			}
		}
		return funcMap, nil
	}
}

func findTemplateFiles(fsys fs.FS, pattern string) (filenames []string, err error) {
	err = fs.WalkDir(fsys, ".",
		func(path string, d fs.DirEntry, err error) error {
			if d.IsDir() || err != nil {
				return err
			}
			if match, e := doublestar.Match(pattern, path); e == nil && match {
				filenames = append(filenames, path)
			}
			return nil
		})
	return
}
