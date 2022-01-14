package swagger

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ghodss/yaml"
	"io/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type OASDocLoader struct {
	path string
	searchFS []fs.FS
}

func newOASDocLoader(path string, searchFS ...fs.FS) *OASDocLoader {
	if len(searchFS) == 0 {
		searchFS = []fs.FS{os.DirFS(".")}
	}
	return &OASDocLoader{
		path:     path,
		searchFS: searchFS,
	}
}

func (l OASDocLoader) Load() (*OpenApiSpec, error) {
	// find docs file
	var file fs.File
	var e error
	for _, fsys := range l.searchFS {
		switch file, e = fsys.Open(l.path); {
		case errors.Is(e, fs.ErrNotExist):
			continue
		case e != nil:
			return nil, e
		}
		break
	}
	if file == nil {
		return nil, fs.ErrNotExist
	}
	defer func() { _ = file.Close() }()

	// load docs
	var oas OpenApiSpec
	switch fileExt := strings.ToLower(path.Ext(l.path)); fileExt {
	case ".yml", ".yaml":
		data, e := ioutil.ReadAll(file)
		if e != nil {
			return nil, e
		}
		if e := yaml.Unmarshal(data, &oas); e != nil {
			return nil, e
		}
	case ".json", ".json5":
		decoder := json.NewDecoder(file)
		if e := decoder.Decode(&oas); e != nil {
			return nil, e
		}
	default:
		return nil, fmt.Errorf("unsupported file extension for OAS document: %s", fileExt)
	}

	return &oas, nil
}
