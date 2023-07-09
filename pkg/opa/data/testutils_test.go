package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils/reflectutils"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"
	"reflect"
)

/**************************
	Test Utils
 **************************/

type CKTestResultHolder struct{}
type TestResultHolder struct {
	Result []policyTarget
}

func AddTestResults(tx *gorm.DB, models ...policyTarget) {
	holder, ok := tx.Statement.Context.Value(CKTestResultHolder{}).(*TestResultHolder)
	if ok {
		holder.Result = append(holder.Result, models...)
	}
}

func ConsumeTestResults(ctx context.Context) []policyTarget {
	holder, ok := ctx.Value(CKTestResultHolder{}).(*TestResultHolder)
	if ok {
		defer func() {
			holder.Result = make([]policyTarget, 0, 4)
		}()
		return holder.Result
	}
	return nil
}

type TestModelTargetExtractor struct{}

// QueryClauses implements schema.QueryClausesInterface,
func (pf TestModelTargetExtractor) QueryClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{NewTestExtractorClause(f, DBOperationFlagRead)}
}

// UpdateClauses implements schema.UpdateClausesInterface,
func (pf TestModelTargetExtractor) UpdateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{NewTestExtractorClause(f, DBOperationFlagUpdate)}
}

// DeleteClauses implements schema.DeleteClausesInterface,
func (pf TestModelTargetExtractor) DeleteClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{NewTestExtractorClause(f, DBOperationFlagDelete)}
}

// CreateClauses implements schema.CreateClausesInterface,
func (pf TestModelTargetExtractor) CreateClauses(f *schema.Field) []clause.Interface {
	return []clause.Interface{NewTestExtractorClause(f, DBOperationFlagCreate)}
}

type TestExtractorClause struct {
	types.NoopStatementModifier
	meta *metadata
	flag DBOperationFlag
}

func NewTestExtractorClause(f *schema.Field, flag DBOperationFlag) clause.Interface {
	meta, e := loadMetadata(f.Schema)
	if e != nil {
		panic(e)
	}
	return &TestExtractorClause{
		meta: meta,
		flag: flag,
	}
}

func (c TestExtractorClause) ModifyStatement(stmt *gorm.Statement) {
	resolved, e := resolvePolicyTargets(stmt, c.meta, c.flag)
	if e != nil {
		_ = stmt.AddError(e)
		return
	}
	// need to make copy of resolved models, because DB operation may change the resolved values and cause false positive results
	for i := range resolved {
		resolved[i].modelValue = c.copyStruct(resolved[i].modelValue)
		if resolved[i].modelValue.IsValid() {
			resolved[i].modelPtr = resolved[i].modelValue.Addr()
		}
	}
	AddTestResults(stmt.DB, resolved...)
	return
}

func (c TestExtractorClause) copyStruct(src reflect.Value) reflect.Value {
	if !src.IsValid() || src.Kind() != reflect.Struct {
		return src
	}
	fields := reflectutils.ListStructField(src.Type(), func(f reflect.StructField) bool {
		return true
	})
	dst := reflect.Indirect(reflect.New(src.Type()))
	for _, f := range fields {
		v := dst.FieldByIndex(f.Index)
		v.Set(src.FieldByIndex(f.Index))
	}
	return dst
}

