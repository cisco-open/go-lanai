package generator

import (
	"github.com/bmatcuk/doublestar/v4"
	"io/fs"
	"os"
)

type DirectoryGenerator struct {
	data       map[string]interface{}
	templateFS fs.FS
	patterns    []string
}

type DirOption struct {
	Option
	Data    map[string]interface{}
	Patterns []string
}

func newDirectoryGenerator(opts ...func(option *DirOption)) *DirectoryGenerator {
	o := &DirOption{}
	for _, fn := range opts {
		fn(o)
	}
	return &DirectoryGenerator{
		data:       o.Data,
		templateFS: o.TemplateFS,
		patterns:    o.Patterns,
	}
}

func (d *DirectoryGenerator) Generate(tmplPath string, tmplInfo fs.FileInfo) error {
	if !tmplInfo.IsDir() || !d.matchPatterns(tmplPath) {
		return nil
	}

	targetDir, err := ConvertSrcRootToTargetDir(tmplPath, d.data)
	if err != nil {
		return err
	}
	logger.Debugf("[Dir] generating %v", targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}

func (d *DirectoryGenerator) matchPatterns(tmplPath string) bool {
	for _, pattern := range d.patterns {
		if match, e := doublestar.Match(pattern, tmplPath); e == nil && match {
			return true
		}
	}
	return false
}
