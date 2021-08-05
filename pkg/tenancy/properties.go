package tenancy

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

func newCacheProperties() *CacheProperties {
	return &CacheProperties{}
}

func bindCacheProperties(ctx *bootstrap.ApplicationContext) CacheProperties {
	props := newCacheProperties()
	if err := ctx.Config().Bind(props, CachePropertiesPrefix); err != nil {
		panic(errors.Wrap(err, "failed to bind CacheProperties"))
	}
	return *props
}