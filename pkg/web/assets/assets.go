package assets

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
)

type assetsMapping struct {
	path string
	root string
}

func New(relativePath string, assetsRootPath string) web.StaticMapping {
	return &assetsMapping{
		path: relativePath,
		root: assetsRootPath,
	}
}

/*****************************
	StaticMapping Interface
******************************/
func (m *assetsMapping) Name() string {
	return m.path
}

func (m *assetsMapping) Path() string {
	return m.path
}

func (m *assetsMapping) Method() string {
	return http.MethodGet
}

func (m *assetsMapping) StaticRoot() string {
	return m.root
}