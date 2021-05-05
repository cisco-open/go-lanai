package cmdutils

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path"
	"text/template"
)

type TemplateOption struct {
	FS         embed.FS
	TmplName   string      // template name
	Output     string      // output path
	OutputPerm os.FileMode // output file permission when create
	Overwrite  bool        // should overwrite if output file already exists
	Model      interface{}
	Customizer func(*template.Template)
}

// GenerateFileWithOption generate file using given FS and template name
func GenerateFileWithOption(ctx context.Context, opt *TemplateOption) error {
	if !path.IsAbs(opt.Output) {
		return fmt.Errorf("template output path should be absolute, but got [%s]", opt.Output)
	}

	// prepare output folder
	dir := path.Dir(opt.Output)
	if e := mkdirIfNotExists(dir); e != nil {
		return fmt.Errorf("unable to create directory of template output [%s]", dir)
	}

	// prepare output file to write, return fast if file already exists and overwrite is not allowed
	if isFileExists(opt.Output) && !opt.Overwrite {
		logger.WithContext(ctx).Infof("file [%s] already exists", opt.Output)
		return nil
	}

	f, e := os.OpenFile(opt.Output, os.O_CREATE|os.O_WRONLY, opt.OutputPerm)
	if e != nil {
		return e
	}
	defer func() {_ = f.Close()}()

	// load template and generate
	t := template.New(opt.TmplName)
	if opt.Customizer != nil {
		opt.Customizer(t)
	}

	t, e = t.ParseFS(opt.FS, opt.TmplName)
	if e != nil {
		return e
	}

	return t.ExecuteTemplate(f, opt.TmplName, opt.Model)
}
