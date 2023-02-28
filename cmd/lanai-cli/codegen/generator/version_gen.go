package generator

import (
	"github.com/getkin/kin-openapi/openapi3"
	"io/fs"
	"path"
	"regexp"
	"sort"
	"text/template"
)

type VersionGenerator struct {
	data       map[string]interface{}
	template   *template.Template
	nameRegex  *regexp.Regexp
	regenRule  string
	filesystem fs.FS
}

const versionGeneratorName = "version"

func newVersionGenerator(opts ...func(option *Option)) *VersionGenerator {
	o := &Option{}
	for _, fn := range opts {
		fn(o)
	}
	rules, ok := o.Rules[versionGeneratorName]
	if ok {
		o.RegenRule = rules.Regeneration
	}

	return &VersionGenerator{
		data:       o.Data,
		template:   o.Template,
		filesystem: o.FS,
		nameRegex:  regexp.MustCompile("^(version.)(.+)(.tmpl)"),
		regenRule:  o.RegenRule,
	}
}

func (m *VersionGenerator) determineFilename(template string) string {
	var result string
	matches := m.nameRegex.FindStringSubmatch(path.Base(template))
	if len(matches) < 2 {
		result = ""
	}

	result = matches[2]
	return result
}

func (m *VersionGenerator) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	if dirEntry.IsDir() || !m.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	// get all versions
	iterateOver := make(map[string][]string)
	for pathName, _ := range m.data[OpenAPIData].(*openapi3.T).Paths {
		version := apiVersion(pathName)
		iterateOver[version] = append(iterateOver[version], pathName)
	}

	var toGenerate []GenerationContext
	for version, versionData := range iterateOver {
		data := copyOf(m.data)
		sort.Strings(versionData)
		data["VersionData"] = versionData
		data["Version"] = version

		targetDir, err := ConvertSrcRootToTargetDir(path.Dir(tmplPath), data, m.filesystem)
		if err != nil {
			return err
		}

		toGenerate = append(toGenerate, *NewGenerationContext(
			tmplPath,
			path.Join(targetDir, m.determineFilename(tmplPath)),
			m.regenRule,
			data,
		))
	}

	for _, gc := range toGenerate {
		logger.Infof("version generator generating %v\n", gc.filename)
		err := GenerateFileFromTemplate(gc, m.template)
		if err != nil {
			return err
		}
	}

	return nil
}
