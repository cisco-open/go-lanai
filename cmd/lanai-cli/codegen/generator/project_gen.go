package generator

import (
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"text/template"
)

const (
	defaultProjectPriorityOrder = 0
	defaultProjectRegex         = "^(?:project.)(.+)(?:.tmpl)"
)

// ProjectGenerator generates 1 file based on the templatePath being used
type ProjectGenerator struct {
	data          map[string]interface{}
	template      *template.Template
	nameRegex     *regexp.Regexp
	filesystem    fs.FS
	priorityOrder int
}

// newProjectGenerator returns a new generator for single files
func newProjectGenerator(opts ...func(option *Option)) *ProjectGenerator {
	o := &Option{}
	for _, fn := range opts {
		fn(o)
	}

	priorityOrder := o.PriorityOrder
	if priorityOrder == 0 {
		priorityOrder = defaultProjectPriorityOrder
	}

	regex := defaultProjectRegex
	if o.Prefix != "" {
		regex = fmt.Sprintf("^(%v)(.+)(.tmpl)", o.Prefix)
	}

	return &ProjectGenerator{
		data:          o.Data,
		template:      o.Template,
		nameRegex:     regexp.MustCompile(regex),
		filesystem:    o.FS,
		priorityOrder: priorityOrder,
	}
}

func (o *ProjectGenerator) determineFilename(template string) string {
	var result string
	matches := o.nameRegex.FindStringSubmatch(path.Base(template))
	if len(matches) < 2 {
		result = ""
	}

	result = matches[1]
	return result
}

func (o *ProjectGenerator) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	if dirEntry.IsDir() || !o.nameRegex.MatchString(path.Base(tmplPath)) {
		// Skip over it
		return nil
	}

	targetDir, err := ConvertSrcRootToTargetDir(path.Dir(tmplPath), o.data, o.filesystem)
	if err != nil {
		return err
	}
	baseFilename := o.determineFilename(tmplPath)

	file := *NewGenerationContext(
		tmplPath,
		path.Join(targetDir, baseFilename),
		o.data,
	)
	logger.Infof("project generator generating %v\n", targetDir)
	return GenerateFileFromTemplate(file, o.template)
}

func (o *ProjectGenerator) PriorityOrder() int {
	return o.priorityOrder
}
