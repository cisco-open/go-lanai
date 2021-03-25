package web

import (
	"embed"
	"errors"
	"github.com/gin-gonic/gin"
	"io/fs"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

type MergedFs struct {
	staticRoot string
	defaultFs  http.FileSystem
	embeddedFs []embed.FS
}

func NewMergedFs(staticRoot string, fileSystem...embed.FS) *MergedFs{
	m := &MergedFs{
		staticRoot: staticRoot,
		defaultFs: gin.Dir(staticRoot, false),
		embeddedFs: fileSystem,
	}
	return m
}

func (m *MergedFs) Open(name string) (file fs.File, err error) {
	file, err = m.defaultFs.Open(name)
	if err == nil {
		return file, err
	}

	if filepath.Separator != '/' && strings.ContainsRune(name, filepath.Separator) {
		return nil, errors.New("invalid character in file path")
	}
	dir := string(m.staticRoot)
	if dir == "" {
		dir = "."
	}
	fullName := filepath.Join(dir, filepath.FromSlash(path.Clean("/"+name)))

	for _, f := range m.embeddedFs {
		file, err = f.Open(fullName)
		if err == nil {
			return file, err
		}
	}
	return nil, &fs.PathError{Op: "open", Path: fullName, Err: fs.ErrNotExist}
}

