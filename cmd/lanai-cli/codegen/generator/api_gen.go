package generator

import (
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"io/fs"
	"path"
	"regexp"
	"strings"
	"text/template"
)

type ApiGenerator struct {
	data             map[string]interface{}
	template         *template.Template
	filesystem       fs.FS
	nameRegex        *regexp.Regexp
	prefix           string
	priorityOrder    int
	defaultRegenRule string
	rules            map[string]string
}

const (
	defaultApiNameRegex = "^(api\\.)(.+)(.tmpl)"
	apiGeneratorName    = "api"
)

var versionRegex = regexp.MustCompile(".+\\/(v\\d+)\\/(.+)")

func newApiGenerator(opts ...func(option *Option)) *ApiGenerator {
	o := &Option{}
	for _, fn := range opts {
		fn(o)
	}
	priorityOrder := o.PriorityOrder
	if priorityOrder == 0 {
		priorityOrder = defaultApiPriorityOrder
	}

	regex := defaultApiNameRegex
	if o.Prefix != "" {
		regex = fmt.Sprintf("^(%v)(.+)(.tmpl)", o.Prefix)
	}
	return &ApiGenerator{
		data:             o.Data,
		template:         o.Template,
		filesystem:       o.FS,
		nameRegex:        regexp.MustCompile(regex),
		priorityOrder:    priorityOrder,
		defaultRegenRule: o.RegenRule,
		rules:            o.Rules,
	}
}

func (m *ApiGenerator) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	if dirEntry.IsDir() || !m.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	iterateOver := m.data[OpenAPIData].(*openapi3.T).Paths
	var toGenerate []GenerationContext
	for pathName, pathData := range iterateOver {
		data := copyOf(m.data)
		data["PathData"] = pathData
		data["PathName"] = pathName
		data["Version"] = apiVersion(pathName)

		baseFilename := filenameFromPath(pathName)
		targetDir, err := ConvertSrcRootToTargetDir(path.Dir(tmplPath), data, m.filesystem)
		if err != nil {
			return err
		}

		outputFile := path.Join(targetDir, baseFilename)
		regenRule, err := getApplicableRegenRules(outputFile, m.rules, m.defaultRegenRule)
		if err != nil {
			return err
		}
		toGenerate = append(toGenerate, *NewGenerationContext(
			tmplPath,
			outputFile,
			regenRule,
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

func filenameFromPath(pathName string) string {
	// Use everything that comes after the version name
	// /my/api/v1/testpath/{scope} -> testpath_scope.go
	parts := versionRegex.FindStringSubmatch(pathName)
	result := pathName
	if len(parts) == 3 {
		result = parts[2] //testpath/{scope}
	}
	result = strings.ReplaceAll(result, "/", "")
	result = strings.ReplaceAll(result, "{", "_")
	result = strings.ReplaceAll(result, "}", "_")

	// Check if last character is a _, if so just drop it
	result = strings.Trim(result, "_") + ".go"
	return result
}

func apiVersion(pathName string) (version string) {
	parts := versionRegex.FindStringSubmatch(pathName)
	if len(parts) == 3 {
		version = parts[1]
	}
	return version
}

func (m *ApiGenerator) PriorityOrder() int {
	return m.priorityOrder
}
