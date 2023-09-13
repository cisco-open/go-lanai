package generator

import (
	"io/fs"
	"os"
)

type DirectoryGenerator struct {
	data       map[string]interface{}
	templateFS fs.FS
	outputFS   fs.FS
}

type DirOption struct {
	Option
	Data map[string]interface{}
}

func newDirectoryGenerator(opts ...func(option *DirOption)) *DirectoryGenerator {
	o := &DirOption{}
	for _, fn := range opts {
		fn(o)
	}
	return &DirectoryGenerator{
		data:       o.Data,
		templateFS: o.TemplateFS,
		outputFS:   o.OutputFS,
	}
}

func (d *DirectoryGenerator) Generate(tmplPath string, tmplInfo fs.FileInfo) error {
	if !tmplInfo.IsDir() {
		return nil
	}

	targetDir, err := ConvertSrcRootToTargetDir(tmplPath, d.data, d.templateFS)
	if err != nil {
		return err
	}
	logger.Infof("directory generator generating %v\n", targetDir)
	if err := os.MkdirAll(targetDir, 0744); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
