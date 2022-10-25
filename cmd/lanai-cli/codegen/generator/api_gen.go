package generator

import (
	"github.com/getkin/kin-openapi/openapi3"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"text/template"
)

type ApiGenerator struct {
	data       map[string]interface{}
	template   *template.Template
	filesystem fs.FS
}

func newApiGenerator(opts ...func(option *Option)) *ApiGenerator {
	o := &Option{}
	for _, fn := range opts {
		fn(o)
	}
	return &ApiGenerator{
		data:       o.Data,
		template:   o.Template,
		filesystem: o.FS,
	}
}

func (m *ApiGenerator) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	if dirEntry.IsDir() || !regexp.MustCompile("^(api.)(.+)(.tmpl)").MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	iterateOver := m.data[OpenAPIData].(*openapi3.T).Paths
	var toGenerate []GenerationContext
	for pathName, pathData := range iterateOver {
		// TODO: Function to determine baseFilename
		baseFilename := strings.ReplaceAll(pathName, "/", "") + ".go"
		targetDir, err := ConvertSrcRootToTargetDir(path.Dir(tmplPath), m.data, m.filesystem)
		if err != nil {
			return err
		}

		filename := path.Join(targetDir, baseFilename)

		data := copyOf(m.data)
		data["PathData"] = pathData
		data["PathName"] = pathName
		toGenerate = append(toGenerate, *NewGenerationContext(
			tmplPath,
			filename,
			data,
		))

	}

	for _, gc := range toGenerate {
		logger.Infof("api generator generating %v\n", gc.filename)
		err := GenerateFileFromTemplate(gc, m.template)
		if err != nil {
			return err
		}
	}

	return nil
}
