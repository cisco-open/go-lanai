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
	gormPluginName     = gormCallbackPrefix + "tracing"
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

// Name implements gorm.Plugin
func (p gormPlugin) Name() string {
	return "tracing"
}

// Initialize implements gorm.Plugin. This function register tracing related callbacks
// Default callbacks can be found at github.com/go-gorm/gorm/callbacks/callbacks.go
func (p gormPlugin) Initialize(db *gorm.DB) error {
	_ = db.Callback().Create().Before(data.GormCallbackBeforeCreate).
		Register(p.cbBeforeName("create"), p.makeBeforeCallback("create"))
	_ = db.Callback().Create().After(data.GormCallbackAfterCreate).
		Register(p.cbAfterName("create"), p.makeAfterCallback("create"))

	_ = db.Callback().Query().Before(data.GormCallbackBeforeQuery).
		Register(p.cbBeforeName("query"), p.makeBeforeCallback("select"))
	_ = db.Callback().Query().After(data.GormCallbackAfterQuery).
		Register(p.cbAfterName("query"), p.makeAfterCallback("select"))

	_ = db.Callback().Update().Before(data.GormCallbackBeforeUpdate).
		Register(p.cbBeforeName("update"), p.makeBeforeCallback("update"))
	_ = db.Callback().Update().After(data.GormCallbackAfterUpdate).
		Register(p.cbAfterName("update"), p.makeAfterCallback("update"))

	_ = db.Callback().Delete().Before(data.GormCallbackBeforeDelete).
		Register(p.cbBeforeName("delete"), p.makeBeforeCallback("delete"))
	_ = db.Callback().Delete().After(data.GormCallbackAfterDelete).
		Register(p.cbAfterName("delete"), p.makeAfterCallback("delete"))

	_ = db.Callback().Row().Before(data.GormCallbackBeforeRow).
		Register(p.cbBeforeName("row"), p.makeBeforeCallback("row"))
	_ = db.Callback().Row().After(data.GormCallbackAfterRow).
		Register(p.cbAfterName("row"), p.makeAfterCallback("row"))

	_ = db.Callback().Raw().Before(data.GormCallbackBeforeRaw).
		Register(p.cbBeforeName("raw"), p.makeBeforeCallback("sql"))
	_ = db.Callback().Raw().After(data.GormCallbackAfterRaw).
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
