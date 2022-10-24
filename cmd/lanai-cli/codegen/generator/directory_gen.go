package generator

import (
	"io/fs"
	"os"
)

type DirectoryGenerator struct {
	data       map[string]interface{}
	filesystem fs.FS
}

func newDirectoryGenerator(opts ...func(option *Option)) *DirectoryGenerator {
	o := &Option{}
	for _, fn := range opts {
		fn(o)
	}
	return &DirectoryGenerator{
		data:       o.Data,
		filesystem: o.FS,
	}
}

func (d *DirectoryGenerator) Generate(tmplPath string, dirEntry fs.DirEntry) error {
	if !dirEntry.IsDir() {
		return nil
	}

	targetDir, err := ConvertSrcRootToTargetDir(tmplPath, d.data, d.filesystem)
	if err != nil {
		return err
	}
	logger.Infof("directory generator generating %v\n", targetDir)
	if err := os.Mkdir(targetDir, 0744); err != nil && !os.IsExist(err) {
		return err
	}

	return nil
}
