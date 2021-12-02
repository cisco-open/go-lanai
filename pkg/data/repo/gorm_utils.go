package repo

import (
	"context"
)

var defaultUtils Utility

// GormUtils implements Utility interface
type GormUtils struct {
	api        GormApi
	model      interface{}
	resolver   GormSchemaResolver
	uniqueness []index
}

func Utils(options ...interface{}) Utility {
	if len(options) == 0 {
		return defaultUtils
	}
	switch factory := globalFactory.(type) {
	case *GormFactory:
		return newGormUtils(factory, options)
	default:
		panic("global repo factory is not set, unable to create Utility")
	}
}

func newGormUtils(factory *GormFactory, options ...interface{}) *GormUtils {
	api := factory.NewGormApi(options...)
	return &GormUtils{
		api:        api,
	}
}

func (g GormUtils) Model(model interface{}) Utility {
	resolver, _ := newGormSchemaResolver(g.api.DB(context.Background()), model)
	return &GormUtils{
		api:      g.api,
		model:    model,
		resolver: resolver,
	}
}

func (g GormUtils) ResolveSchema(ctx context.Context, model interface{}) (SchemaResolver, error) {
	return newGormSchemaResolver(g.api.DB(ctx), model)
}

func (g GormUtils) CheckUniqueness(ctx context.Context, v interface{}, keys ...interface{}) (dups map[string]interface{}, err error) {
	resolver, e := g.getSchemaResolver(ctx, v)
	if e != nil {
		return nil, e
	}
	return gormCheckUniqueness(ctx, g.api, resolver, v, keys)
}

func (g GormUtils) getSchemaResolver(ctx context.Context, v interface{}) (GormSchemaResolver, error) {
	switch {
	case g.resolver != nil :
		return g.resolver, nil
	default:
		return newGormSchemaResolver(g.api.DB(ctx), v)
	}
}