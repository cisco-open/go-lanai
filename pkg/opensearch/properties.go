package opensearch

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"embed"
	"github.com/opensearch-project/opensearch-go"
	"github.com/pkg/errors"
)

const (
	PropertiesPrefix = "data.opensearch"
)

//go:embed defaults-opensearch.yml
var defaultConfigFS embed.FS

type Properties struct {
	Addresses []string `json:"addresses"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	CACert    string   `json:""`
}

func NewOpenSearchProperties() *Properties {
	return &Properties{
		Addresses: []string{"http://localhost:9200"},
		Username:  "admin",
		Password:  "admin",
	}
}

func BindOpenSearchProperties(ctx *bootstrap.ApplicationContext) Properties {
	props := NewOpenSearchProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind OpenSearchProperties"))
	}
	return *props
}

func (c Properties) GetConfig() opensearch.Config {
	return opensearch.Config{
		Addresses: c.Addresses,
		Username:  c.Username,
		Password:  c.Password,
	}
}
