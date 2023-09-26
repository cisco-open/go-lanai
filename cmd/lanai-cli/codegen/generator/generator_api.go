package generator

import (
	"context"
	"fmt"
	"github.com/getkin/kin-openapi/openapi3"
	"io/fs"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

type ApiGenerator struct {
	data             map[string]interface{}
	template         *template.Template
	templateFS       fs.FS
	matcher          TemplateMatcher
	outputResolver   TemplateOutputResolver
	order            int
	defaultRegenRule RegenMode
	rules            RegenRules
}

const (
	apiDefaultPrefix        = "api"
	apiStructPrefix         = "api-struct"
)

var versionRegex = regexp.MustCompile(".+\\/(v\\d+)\\/(.+)")

type ApiOption struct {
	GeneratorOption
	Matcher        TemplateMatcher
	Prefix         string
	Order          int
}

func newApiGenerator(gOpt GeneratorOption, opts ...func(opt *ApiOption)) *ApiGenerator {
	o := &ApiOption{
		GeneratorOption: gOpt,
		Prefix:          apiDefaultPrefix,
	}
	for _, fn := range opts {
		fn(o)
	}

	if o.Matcher == nil {
		pattern := fmt.Sprintf(patternWithFilePrefix, o.Prefix)
		o.Matcher = isTmplFile().And(matchPatterns(pattern))
	}

	return &ApiGenerator{
		data:             o.Data,
		template:         o.Template,
		templateFS:       o.TemplateFS,
		matcher:          o.Matcher,
		outputResolver:   apiOutputResolver(),
		order:            o.Order,
		defaultRegenRule: o.DefaultRegenMode,
		rules:            o.RegenRules,
	}
}

func (g *ApiGenerator) Generate(ctx context.Context, tmplDesc TemplateDescriptor) error {
	if ok, e := g.matcher.Matches(tmplDesc); e != nil || !ok {
		return e
	}

	iterateOver := g.data[KDataOpenAPI].(*openapi3.T).Paths
	var toGenerate []GenerationContext
	for pathName, pathData := range iterateOver {
		data := copyOf(g.data)
		data["PathData"] = pathData
		data["PathName"] = pathName
		data["Version"] = apiVersion(pathName)

		output, e := g.outputResolver.Resolve(ctx, tmplDesc, data)
		if e != nil {
			return e
		}

		regenRule, err := getApplicableRegenRules(output.Path, g.rules, g.defaultRegenRule)
		if err != nil {
			return err
		}
		toGenerate = append(toGenerate, *NewGenerationContext(
			tmplDesc.Path,
			output.Path,
			regenRule,
			data,
		))
	}

	for _, gc := range toGenerate {
		logger.Debugf("[API] generating %v", gc.filename)
		err := GenerateFileFromTemplate(gc, g.template)
		if err != nil {
			return err
		}
		globalCounter.Record(gc.filename)
	}
	return nil
}

func (g *ApiGenerator) Order() int {
	return g.order
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

func apiOutputResolver() TemplateOutputResolver {
	return TemplateOutputResolverFunc(func(ctx context.Context, tmplDesc TemplateDescriptor, data GenerationData) (TemplateOutputDescriptor, error) {
		path, e := ConvertSrcRootToTargetDir(tmplDesc.Path, data)
		if e != nil {
			return TemplateOutputDescriptor{}, e
		}

		dir := filepath.Dir(path)
		filename := filenameFromPath(data["PathName"].(string))
		return TemplateOutputDescriptor{
			Path: filepath.Join(dir, filename),
		}, nil
	})
}

