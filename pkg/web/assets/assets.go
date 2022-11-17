package assets

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/web"
	"net/http"
)

type assetsMapping struct {
	path    string
	root    string
	aliases map[string]string
}

func New(relativePath string, assetsRootPath string) web.StaticMapping {
	return &assetsMapping{
		path: relativePath,
		root: assetsRootPath,
		aliases: map[string]string{},
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

func (m *assetsMapping) Aliases() map[string]string {
	return m.aliases
}

func (m *assetsMapping) AddAlias(path, filePath string) web.StaticMapping {
	m.aliases[path] = filePath
	return m
}