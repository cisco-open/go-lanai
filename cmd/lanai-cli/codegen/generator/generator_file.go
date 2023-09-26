package generator

import (
	"context"
	"fmt"
	"io/fs"
	"regexp"
	"text/template"
)

const (
	fileDefaultPrefix   = "project"
)

// FileGenerator is a basic generator that generates 1 file based on the templatePath being used
type FileGenerator struct {
	data             map[string]interface{}
	template         *template.Template
	templateFS       fs.FS
	matcher          TemplateMatcher
	outputResolver   TemplateOutputResolver
	order            int
	defaultRegenRule RegenMode
	rules            RegenRules
}

type FileOption struct {
	GeneratorOption
	Matcher        TemplateMatcher
	OutputResolver TemplateOutputResolver
	Prefix         string
	Order          int
}

// newFileGenerator returns a new generator for single files
func newFileGenerator(gOpt GeneratorOption, opts ...func(opt *FileOption)) *FileGenerator {
	o := &FileOption{
		GeneratorOption: gOpt,
		Prefix:          fileDefaultPrefix,
	}
	for _, fn := range opts {
		fn(o)
	}

	if o.Matcher == nil {
		pattern := fmt.Sprintf(patternWithFilePrefix, o.Prefix)
		o.Matcher = isTmplFile().And(matchPatterns(pattern))
	}

	if o.OutputResolver == nil {
		regex := fmt.Sprintf(outputRegexWithFilePrefix, regexp.QuoteMeta(o.Prefix))
		o.OutputResolver = regexOutputResolver(regex)
	}

	logger.Debugf("Templates [%v] DefaultRegenMode: %v", o.Matcher, o.DefaultRegenMode)
	return &FileGenerator{
		data:             o.Data,
		template:         o.Template,
		matcher:          o.Matcher,
		outputResolver:   o.OutputResolver,
		templateFS:       o.TemplateFS,
		order:            o.Order,
		defaultRegenRule: o.DefaultRegenMode,
		rules:            o.RegenRules,
	}
}

func (g *FileGenerator) Generate(ctx context.Context, tmplDesc TemplateDescriptor) error {
	if ok, e := g.matcher.Matches(tmplDesc); e != nil || !ok {
		return e
	}

	output, e := g.outputResolver.Resolve(ctx, tmplDesc, g.data)
	if e != nil {
		return e
	}

	regenRule, err := getApplicableRegenRules(output.Path, g.rules, g.defaultRegenRule)
	if err != nil {
		return err
	}
	gc := *NewGenerationContext(
		tmplDesc.Path,
		output.Path,
		regenRule,
		g.data,
	)
	logger.Debugf("[File] generating %v", gc.filename)
	if e := GenerateFileFromTemplate(gc, g.template); e != nil {
		return e
	}
	globalCounter.Record(output.Path)
	return nil
}

func (g *FileGenerator) Order() int {
	return g.order
}
