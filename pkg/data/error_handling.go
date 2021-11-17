package data

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/order"
	"fmt"
	"go.uber.org/fx"
	"gorm.io/gorm"
	"sort"
)

const (
	gormCallbackPrefix = "lanai:error:"
	gormPluginName = gormCallbackPrefix + "translate"
)

// errorHandlingGormConfigurer implement a GormConfigurer that installs errorTranslatorGormPlugin for error handling/transformation
// see errorTranslatorGormPlugin for more details
type errorHandlingGormConfigurer []ErrorTranslator

func ErrorHandlingGormConfigurer() fx.Annotated {
	return fx.Annotated{
		Group:  GormConfigurerGroup,
		Target: newErrHandlingGormConfigurer,
	}
}

type ehDI struct {
	fx.In
	Translators []ErrorTranslator `group:"gorm_config"`
}

func newErrHandlingGormConfigurer(di ehDI) GormConfigurer {
	return errorHandlingGormConfigurer(di.Translators)
}

func (c errorHandlingGormConfigurer) Order() int {
	return 0
}

func (c errorHandlingGormConfigurer) Configure(config *gorm.Config) {
	if config.Plugins == nil {
		config.Plugins = map[string]gorm.Plugin{}
	}
	config.Plugins[gormPluginName] = newErrorHandlingGormPlugin(c...)
}

// errorTranslatorGormPlugin installs gorm callbacks of all operations and give ErrorTranslator a chance to handle errors
// before *gorm.DB operations return
type errorTranslatorGormPlugin []ErrorTranslator

func newErrorHandlingGormPlugin(translators ...ErrorTranslator) errorTranslatorGormPlugin {
	sort.SliceStable(translators, func(i, j int) bool {
		return order.OrderedFirstCompare(translators[i], translators[j])
	})
	return translators
}

func (errorTranslatorGormPlugin) Name() string {
	return gormPluginName
}

func (p errorTranslatorGormPlugin) Initialize(db *gorm.DB) error {
	errs := map[string]error{}
	cbName := gormCallbackPrefix + "translate"
	errs["Create"] = db.Callback().Create().After("*").Register(cbName, p.translateErrorCallback())
	errs["Query"] = db.Callback().Query().After("*").Register(cbName, p.translateErrorCallback())
	errs["Update"] = db.Callback().Update().After("*").Register(cbName, p.translateErrorCallback())
	errs["Delete"] = db.Callback().Delete().After("*").Register(cbName, p.translateErrorCallback())
	errs["Raw"] = db.Callback().Raw().After("*").Register(cbName, p.translateErrorCallback())
	errs["Row"] = db.Callback().Row().After("*").Register(cbName, p.translateErrorCallback())

	for k, e := range errs {
		if e != nil {
			return fmt.Errorf("unable to install error transformation callbacks for %s: %v", k, e)
		}
	}
	return nil
}

func (p errorTranslatorGormPlugin) translateErrorCallback() func(*gorm.DB) {
	return func(db *gorm.DB) {
		if db.Error == nil {
			return
		}
		for _, translator := range p {
			db.Error = translator.Translate(db.Statement.Context, db.Error)
			if db.Error == nil {
				return
			}
		}
	}
}

