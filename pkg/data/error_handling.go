// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package data

import (
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/utils/order"
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
			switch t := translator.(type) {
			case GormErrorTranslator:
				db.Error = t.TranslateWithDB(db)
			default:
				db.Error = translator.Translate(db.Statement.Context, db.Error)
			}
			if db.Error == nil {
				return
			}
		}
	}
}

