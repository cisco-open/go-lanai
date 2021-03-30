package tenant_hierarchy_accessor

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/bootstrap"
	"github.com/pkg/errors"
)

/***********************
	Cache
************************/
const CachePropertiesPrefix = "security.cache"

type CacheProperties struct {
	DbIndex int `json:"db-index"`
}

func NewCacheProperties() *CacheProperties {
	return &CacheProperties{}
}

func BindCacheProperties(ctx *bootstrap.ApplicationContext) CacheProperties {
	props := NewCacheProperties()
	if err := ctx.Config().Bind(props, CachePropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CacheProperties"))
	}
	return *props
}