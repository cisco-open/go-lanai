package generator

import (
	"context"
	"io/fs"
	"os"
)

type DirectoryGenerator struct {
	data       map[string]interface{}
	templateFS fs.FS
	matcher        TemplateMatcher
	outputResolver TemplateOutputResolver
}

type DirOption struct {
	GeneratorOption
	Matcher TemplateMatcher
	OutputResolver TemplateOutputResolver
}

func newDirectoryGenerator(gOpt GeneratorOption, opts ...func(option *DirOption)) *DirectoryGenerator {
	o := &DirOption{
		GeneratorOption: gOpt,
		Matcher: isDir(),
		OutputResolver: regexOutputResolver(""),
	}
	for _, fn := range opts {
		fn(o)
	}
	return &DirectoryGenerator{
		data:           o.Data,
		templateFS:     o.TemplateFS,
		matcher:        o.Matcher,
		outputResolver: o.OutputResolver,
	}
}

func (d *DirectoryGenerator) Generate(ctx context.Context, tmplDesc TemplateDescriptor) error {
	if ok, e := d.matcher.Matches(tmplDesc); e != nil || !ok {
		return e
	}

	output, e := d.outputResolver.Resolve(ctx, tmplDesc, d.data)
	if e != nil {
		return e
	}

	logger.Debugf("[Dir] generating %v", output.Path)
	if err := os.MkdirAll(output.Path, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
