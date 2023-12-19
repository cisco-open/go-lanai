package opensearch

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/certs"
	"embed"
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
	TLS       TLS      `json:"tls"`
}

type TLS struct {
	Enable bool                   `json:"enable"`
	Certs  certs.SourceProperties `json:"certs"`
}

func NewOpenSearchProperties() *Properties {
	return &Properties{} // None by default, they should all be defined in the defaults-opensearch.yml
}

func BindOpenSearchProperties(ctx *bootstrap.ApplicationContext) *Properties {
	props := NewOpenSearchProperties()
	if err := ctx.Config().Bind(props, PropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind OpenSearchProperties"))
	}
	return props
}
