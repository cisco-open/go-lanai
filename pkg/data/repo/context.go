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

package repo

import (
	"context"
	"github.com/cisco-open/go-lanai/pkg/data"
	"reflect"
)

var (
	ErrorInvalidCrudModel = data.NewDataError(data.ErrorCodeInvalidCrudModel, "invalid model for CrudRepository")
	ErrorUnsupportedOptions = data.NewDataError(data.ErrorCodeUnsupportedOptions, "unsupported Option")
	ErrorUnsupportedCondition = data.NewDataError(data.ErrorCodeUnsupportedCondition, "unsupported Condition")
	ErrorInvalidCrudParam = data.NewDataError(data.ErrorCodeInvalidCrudParam, "invalid CRUD param")
	ErrorInvalidPagination = data.NewDataError(data.ErrorCodeInvalidPagination, "invalid pagination")
	ErrorInvalidUtilityUsage = data.NewDataError(data.ErrorCodeInvalidApiUsage, "invalid utility usage")
)

// Factory usually used in repository creation.
type Factory interface {
	// NewCRUD create an implementation specific CrudRepository.
	// "model" represent the model this repository works on. It could be Struct or *Struct
	// It panic if model is not a valid model definition
	// accepted options depends on implementation. for gorm, *gorm.Session can be supplied
	NewCRUD(model interface{}, options...interface{}) CrudRepository
}

// Condition is typically used for generic CRUD repository
// supported condition depends on operation and underlying implementation:
// 	- map[string]interface{} (should be generally supported)
//		e.g. {"col1": "val1", "col2": 10} -> "WHERE col1 = "val1" AND col2 = 10"
//	- Struct (supported by gorm)
//		e.g. &User{FirstName: "john", Age: 20} -> "WHERE first_name = "john" AND age = 20"
//  - raw condition generated by Where
//  - scope function func(*gorm.DB) *gorm.DB
//  - valid gorm clause.
// 		e.g. clause.Where
//  - slice of above
// If given condition is not supported, an error with code data.ErrorCodeUnsupportedCondition will be return
//  - TODO 1: more features leveraging "gorm" lib. Ref: https://gorm.io/docs/query.html#Conditions
//  - TODO 2: more detailed documentation of already supported types
type Condition interface{}

// Option is typically used for generic CRUD repository
// supported options depends on operation and underlying implementation
//  - Omit for read/write
//  - Joins for read
//  - Preload for read
//  - Select for read/write
// 	- Sort for read
//  - Page for read
//  - func(*gorm.DB) *gormDB
// 	- ...
// If given condition is not supported, an error with code data.ErrorCodeUnsupportedOptions will be return
// TODO Provide more supporting features
type Option interface {}

// SchemaResolver resolves schema related values
type SchemaResolver interface{
	// ModelType returns reflect type of the model
	ModelType() reflect.Type
	// ModelName returns the name of the model
	ModelName() string
	// Table resolve table name of the model
	Table() string
	// ColumnName resolves the column name by given field name of the Model.
	// field path is supported, e.g. "AssociationField.FieldName"
	ColumnName(fieldName string) string
	// ColumnDataType resolves the column data type string by given field name of the Model
	// field path is supported, e.g. "AssociationField.FieldName"
	ColumnDataType(fieldName string) string
	// RelationshipSchema returns SchemaResolver of the relationship fields with given name.
	// This function returns nil if given field name is not a relationship field.
	RelationshipSchema(fieldName string) SchemaResolver
}

type CrudRepository interface {
	SchemaResolver

	// FindById fetch model by primary key and scan it into provided interface.
	// Accepted "dest" types:
	//		*ModelStruct
	FindById(ctx context.Context, dest interface{}, id interface{}, options...Option) error

	// FindAll fetch all model scan it into provided slice.
	// Accepted "dest" types:
	//		*[]*ModelStruct
	//		*[]ModelStruct
	FindAll(ctx context.Context, dest interface{}, options...Option) error

	// FindOneBy fetch single model with given condition and scan result into provided value.
	// Accepted "dest" types:
	//		*ModelStruct
	FindOneBy(ctx context.Context, dest interface{}, condition Condition, options...Option) error

	// FindAllBy fetch all model with given condition and scan result into provided value.
	// Accepted "dest" types:
	//		*[]*ModelStruct
	//		*[]ModelStruct
	FindAllBy(ctx context.Context, dest interface{}, condition Condition, options...Option) error

	// CountAll counts all
	CountAll(ctx context.Context, options...Option) (int, error)

	// CountBy counts based on conditions.
	CountBy(ctx context.Context, condition Condition, options...Option) (int, error)

	// Save create or update model or model array.
	// Accepted "v" types:
	//		*ModelStruct
	//		[]*ModelStruct
	//		[]ModelStruct
	//		ModelStruct
	//  	map[string]interface{}
	// Note:
	//		1. map[string]interface{} might not be supported by underlying implementation
	//		2. ModelStruct is not recommended because auto-generated field default will be lost
	Save(ctx context.Context, v interface{}, options...Option) error

	// Create create model or model array. returns error if model already exists
	// Accepted "v" types:
	//		*ModelStruct
	//		[]*ModelStruct
	//		[]ModelStruct
	//		ModelStruct
	//  	map[string]interface{}
	// Note:
	//		1. map[string]interface{} might not be supported by underlying implementation
	//		2. ModelStruct is not recommended because auto-generated field default will be lost
	Create(ctx context.Context, v interface{}, options...Option) error

	// Update update model, only non-zero fields of "v" are updated.
	// "model" is the model to be updated, loaded from DB
	// Accepted "dest" types:
	//		*ModelStruct
	//		ModelStruct
	// Accepted "v" types:
	//		ModelStruct
	//		*ModelStruct
	//		map[string]interface{}
	// Update might support Select and Omit depends on implementation
	// Note: when ModelStruct or *ModelStruct is used, GORM limitation applys:
	// 		 https://gorm.io/docs/update.html#Updates-multiple-columns
	// The workaround is to use Select as described here:
	//		 https://gorm.io/docs/update.html#Update-Selected-Fields
	Update(ctx context.Context, model interface{}, v interface{}, options...Option) error

	// Delete delete given model or model array
	// Accepted "v" types:
	//		*ModelStruct
	//		[]*ModelStruct
	//		[]ModelStruct
	//		ModelStruct
	// returns error if such deletion violate any existing foreign key constraints
	Delete(ctx context.Context, v interface{}, options...Option) error

	// DeleteBy delete models matching given condition.
	// returns error if such deletion violate any existing foreign key constraints
	DeleteBy(ctx context.Context, condition Condition, options...Option) error

	// Truncate attempt to truncate the table associated the repository
	// returns error if such truncattion violate any existing foreign key constraints
	// Warning: User with Caution: interface is not finalized
	Truncate(ctx context.Context) error
}

// Utility is a collection of repository related patterns that are useful for common service layer implementation
type Utility interface{

	// Model returns a model-specific implementation of Utility.
	// It is useful when model type cannot be deduced from provided parameters
	// e.g.
	// <code>
	//	// this is ok, because value &SomeModel{} also imply the model type
	// 	Utility.CheckUniqueness(ctx, &SomeModel{})
	//
	//	// this is ok, because the model type is specified via Model()
	// 	Utility.Model(&SomeModel{}).CheckUniqueness(ctx, map[string]interface{}{"Field1":"Value1"})
	//
	//  // this is NOT ok, because there is no way to tell what model it is
	// 	Utility.CheckUniqueness(ctx, map[string]interface{}{"Field1":"Value1"})
	// </code>
	Model(model interface{}) Utility

	// ResolveSchema parse given model and returns its SchemaResolver
	// It's highly recommended to use CrudRepository instead of this one.
	ResolveSchema(ctx context.Context, model interface{}) (SchemaResolver, error)

	// CheckUniqueness check if any non-zero unique field of given model ("v") violate unique key constraints in DB
	// When uniqueness check fails, the returned map contains field names and values that violate the constraints
	// and a data.ErrorSubTypeDataIntegrity error
	// Accepted "v" types:
	// 		*ModelStruct
	//		[]*ModelStruct
	// 		map[string]interface{} where key is model's field name or col name
	// By default, this function would use models' schema to figure out unique keys.
	// However, if "keys" is provided, it would override schema definition
	// Supported "keys types:
	// 		string: single field/column
	// 		[]string: index key
	// Note: primary key is not included by default
	CheckUniqueness(ctx context.Context, v interface{}, keys...interface{}) (map[string]interface{}, error)
}
