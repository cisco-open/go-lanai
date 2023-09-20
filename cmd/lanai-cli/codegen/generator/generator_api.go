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
	templateFS       fs.FS
	nameRegex        *regexp.Regexp
	prefix           string
	order            int
	defaultRegenRule RegenMode
	rules            RegenRules
}

const (
	apiMatcherRegexTemplate = `^(%s)(.+)(.tmpl)`
	apiDefaultPrefix        = "api."
	apiStructDefaultPrefix  = "api-struct."
)

var versionRegex = regexp.MustCompile(".+\\/(v\\d+)\\/(.+)")

type ApiOption struct {
	Option
	Template *template.Template
	Data     map[string]interface{}
	Prefix   string
	Order    int
}

func newApiGenerator(opts ...func(opt *ApiOption)) *ApiGenerator {
	o := &ApiOption{
		Prefix: apiDefaultPrefix,
		Order:  defaultApiPriorityOrder,
	}
	for _, fn := range opts {
		fn(o)
	}

	regex := fmt.Sprintf(apiMatcherRegexTemplate, regexp.QuoteMeta(o.Prefix))
	return &ApiGenerator{
		data:             o.Data,
		template:         o.Template,
		templateFS:       o.TemplateFS,
		nameRegex:        regexp.MustCompile(regex),
		order:            o.Order,
		defaultRegenRule: o.DefaultRegenMode,
		rules:            o.RegenRules,
	}
}

func (m *ApiGenerator) Generate(tmplPath string, tmplInfo fs.FileInfo) error {
	if tmplInfo.IsDir() || !m.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	iterateOver := m.data[KDataOpenAPI].(*openapi3.T).Paths
	var toGenerate []GenerationContext
	for pathName, pathData := range iterateOver {
		data := copyOf(m.data)
		data["PathData"] = pathData
		data["PathName"] = pathName
		data["Version"] = apiVersion(pathName)

		baseFilename := filenameFromPath(pathName)
		targetDir, err := ConvertSrcRootToTargetDir(path.Dir(tmplPath), data)
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
		logger.Debugf("[API] generating %v", gc.filename)
		err := GenerateFileFromTemplate(gc, m.template)
		if err != nil {
			return err
		}
		globalCounter.Record(gc.filename)
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

func (m *ApiGenerator) Order() int {
	return m.order
}
