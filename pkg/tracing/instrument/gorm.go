package instrument

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/tracing"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"go.uber.org/fx"
	"gorm.io/gorm"
)

const (
	gormCallbackPrefix = "lanai:"
	gormPluginName = gormCallbackPrefix + "tracing"
)

const (
	gormCbBeforeCreate = "gorm:before_create"
	gormCbAfterCreate  = "gorm:after_create"
	gormCbBeforeQuery  = "gorm:query"
	gormCbAfterQuery   = "gorm:after_query"
	gormCbBeforeUpdate = "gorm:before_update"
	gormCbAfterUpdate  = "gorm:after_update"
	gormCbBeforeDelete = "gorm:before_delete"
	gormCbAfterDelete  = "gorm:after_delete"
	gormCbBeforeRow    = "gorm:row"
	gormCbAfterRow     = "gorm:row"
	gormCbBeforeRaw    = "gorm:raw"
	gormCbAfterRaw     = "gorm:raw"
)

type gormConfigurer struct {
	tracer opentracing.Tracer
}

func GormTracingProvider() fx.Annotated {
	return fx.Annotated{
		Group:  data.GormConfigurerGroup,
		Target: newGormTracingConfigurer,
	}
}

func newGormTracingConfigurer(tracer opentracing.Tracer) data.GormConfigurer {
	return &gormConfigurer{
		tracer: tracer,
	}
}

func (c gormConfigurer) Order() int {
	return order.Highest + 1
}

func (c gormConfigurer) Configure(config *gorm.Config) {
	if config.Plugins == nil {
		config.Plugins = map[string]gorm.Plugin{}
	}
	config.Plugins[gormPluginName] = &gormPlugin{
		tracer: c.tracer,
	}
}

type gormCallbackFunc func(*gorm.DB)

type gormPlugin struct {
	tracer opentracing.Tracer
}

func (p gormPlugin) Name() string {
	return "tracing"
}

func (p gormPlugin) Initialize(db *gorm.DB) (err error) {
	err = db.Callback().Create().Before(gormCbBeforeCreate).
		Register(p.cbBeforeName("create"), p.makeBeforeCallback("create"))
	err = db.Callback().Create().After(gormCbAfterCreate).
		Register(p.cbAfterName("create"), p.makeAfterCallback("create"))

	err = db.Callback().Query().Before(gormCbBeforeQuery).
		Register(p.cbBeforeName("query"), p.makeBeforeCallback("select"))
	err = db.Callback().Query().After(gormCbAfterQuery).
		Register(p.cbAfterName("query"), p.makeAfterCallback("select"))

	err = db.Callback().Update().Before(gormCbBeforeUpdate).
		Register(p.cbBeforeName("update"), p.makeBeforeCallback("update"))
	err = db.Callback().Update().After(gormCbAfterUpdate).
		Register(p.cbAfterName("update"), p.makeAfterCallback("update"))

	err = db.Callback().Delete().Before(gormCbBeforeDelete).
		Register(p.cbBeforeName("delete"), p.makeBeforeCallback("delete"))
	err = db.Callback().Delete().After(gormCbAfterDelete).
		Register(p.cbAfterName("delete"), p.makeAfterCallback("delete"))

	err = db.Callback().Row().Before(gormCbBeforeRow).
		Register(p.cbBeforeName("row"), p.makeBeforeCallback("row"))
	err = db.Callback().Row().After(gormCbAfterRow).
		Register(p.cbAfterName("row"), p.makeAfterCallback("row"))

	err = db.Callback().Raw().Before(gormCbBeforeRaw).
		Register(p.cbBeforeName("raw"), p.makeBeforeCallback("sql"))
	err = db.Callback().Raw().After(gormCbAfterRaw).
		Register(p.cbAfterName("raw"), p.makeAfterCallback("sql"))

	return nil
}

func (p gormPlugin) makeBeforeCallback(opName string) gormCallbackFunc {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		name := tracing.OpNameDB + " " + opName
		table := db.Statement.Table
		if db.Statement.TableExpr != nil {
			table = db.Statement.TableExpr.SQL
		}

		opts := []tracing.SpanOption{
			tracing.SpanKind(ext.SpanKindRPCClientEnum),
			tracing.SpanTag("table", table),
		}

		db.Statement.Context = tracing.WithTracer(p.tracer).
			WithOpName(name).
			WithOptions(opts...).
			DescendantOrNoSpan(ctx)
	}
}

func (p gormPlugin) makeAfterCallback(_ string) gormCallbackFunc {
	return func(db *gorm.DB) {
		ctx := db.Statement.Context
		op := tracing.WithTracer(p.tracer)
		if db.Error != nil {
			op = op.WithOptions(tracing.SpanTag("err", db.Error))
		} else {
			op = op.WithOptions(tracing.SpanTag("rows", db.RowsAffected))
		}
		db.Statement.Context = op.FinishAndRewind(ctx)
	}
}

func (p gormPlugin) cbBeforeName(name string) string {
	return gormCallbackPrefix + "before_" + name
}

func (p gormPlugin) cbAfterName(name string) string {
	return gormCallbackPrefix + "after_" + name
}