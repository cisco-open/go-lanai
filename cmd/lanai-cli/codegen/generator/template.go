package generator

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/cmd/lanai-cli/codegen/generator/internal"
	"io/fs"
	"path"
	"text/template"
)

func GenerateFiles(filesystem fs.FS, opts ...func(*Option)) error {
	generators := NewGenerators(opts...)
	if err := fs.WalkDir(filesystem, ".",
		func(p string, d fs.DirEntry, err error) error {
			// load the files into the generator
			generators.Load(p, d)
			return nil
		}); err != nil {
		return err
	}
	//	 generate
	return generators.Generate()
}

type LoaderOptions struct {
	InitialRegexes map[string]string
}

func LoadTemplates(filesystem fs.FS, loaderOptions LoaderOptions) (*template.Template, error) {
	tmpl := template.New("templates")
	tmpl.Funcs(templateFunctions())

	internal.Load()
	internal.AddPredefinedRegexes(loaderOptions.InitialRegexes)
	if err := fs.WalkDir(filesystem, ".",
		func(p string, d fs.DirEntry, err error) error {
			if !d.IsDir() && isTemplateFile(d) {
				content, err := fs.ReadFile(filesystem, p)
				if err != nil {
					return err
				} else if content == nil {
					return nil
				}

				tmpl, err = tmpl.New(p).Parse(string(content))
				return err
			}
			return nil
		}); err != nil {
		return nil, err
	}
	return tmpl, nil
}

func templateFunctions() template.FuncMap {
	templateFunctions := make(template.FuncMap)
	for _, fm := range internal.TemplateFuncMaps {
		for k, v := range fm {
			templateFunctions[k] = v
		}
	}
	return templateFunctions
}

func isTemplateFile(d fs.DirEntry) bool {
	return path.Ext(d.Name()) == ".tmpl"
}
