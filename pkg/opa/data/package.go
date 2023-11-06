package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
	"gorm.io/gorm"
	"reflect"
	"strings"
)

var logger = log.New("OPA.Data")

/**********************
	Global Functions
 **********************/

// ResolveResource parse given model and resolve resource type and resource values using "opa" tags
// Typically used together with opa.AllowResource as manual policy enforcement.
// ModelType should be model struct with PolicyFilter and valid "opa" tags.
// Note: resValues could be nil if all OPA related values are zero
func ResolveResource[ModelType any](model *ModelType) (resType string, resValues *opa.ResourceValues, err error) {
	rv := reflect.ValueOf(model)
	if rv.Kind() != reflect.Pointer || rv.Elem().Kind() != reflect.Struct {
		return "", nil, ErrUnsupportedUsage.WithMessage(`unable to resolve metadata of "%T": model need to be a struct`, model)
	}

	var meta *Metadata
	if meta, err = resolveMetadata(model); err != nil {
		return "", nil, ErrUnsupportedUsage.WithMessage(`unable to resolve metadata of "%T": %v`, model, err)
	}

	resType = meta.ResType
	target := policyTarget{
		meta:       meta,
		modelPtr:   rv,
		modelValue: rv.Elem(),
		model:      model,
	}

	if resValues, err = target.toResourceValues(); err != nil {
		return "", nil, ErrUnsupportedUsage.WithMessage(`unable to extract OPA resource values of "%T": %v`, model, err)
	}
	return
}

/********************
	GORM Scopes
 ********************/

// SkipPolicyFiltering is used as a scope for gorm.DB to skip policy-based data filtering
// e.g. db.WithContext(ctx).Scopes(SkipPolicyFiltering()).Find(...)
// Using this scope without context would panic
func SkipPolicyFiltering() func(*gorm.DB) *gorm.DB {
	return FilterByPolicies(0)
}

// FilterByPolicies is used as a scope for gorm.DB to override policy-based data filtering.
// The specified operations are enabled, and the rest are disabled
// e.g. db.WithContext(ctx).Scopes(FilterByPolicies(DBOperationFlagRead)).Find(...)
// Using this scope without context would panic
func FilterByPolicies(flags ...DBOperationFlag) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("FilterByPolicies scope is used without context")
		}
		var mode policyMode
		for _, flag := range flags {
			mode = mode | policyMode(flag)
		}
		ctx := context.WithValue(tx.Statement.Context, ckFilterMode{}, mode)
		tx.Statement.Context = ctx
		return tx
	}
}

// FilterByQueries is used as a scope for gorm.DB to override policy-based data filtering.
// Used to customize queries of specified DB operation. Additional DBOperationFlag-string pairs can be provided.
// e.g. db.WithContext(ctx).Scopes(FilterByQueries(DBOperationFlagRead, "resource.type.allow_read")).Find(...)
// Important: This scope accept FULL QUERY including policy package.
// Notes:
// - It's recommended to use dotted format without leading "data.". PolicyFilter would adjust the format based on operation.
// 	 e.g. "resource.type.allow_read"
// - This scope doesn't enable/disable data-filtering. It only overrides queries set in tag.
// - Using this scope without context would panic
// - Having incorrect parameters cause panic
func FilterByQueries(op DBOperationFlag, query string, more ...interface{}) func(*gorm.DB) *gorm.DB {
	policies := map[DBOperationFlag]string{op: query}
	for i := range more {
		if op, ok := more[i].(DBOperationFlag); ok && i + 1 < len(more) {
			if v, ok := more[i+1].(string); !ok {
				panic("FilterByQueries scope only support DBOperationFlag and string pairs")
			} else if len(v) != 0 {
				policies[op] = v
			}
		}
		i++
	}
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("FilterByQueries scope is used without context")
		}
		ctx := tx.Statement.Context
		existing, ok := ctx.Value(ckFilterQueries{}).(map[DBOperationFlag]string)
		if !ok {
			existing = map[DBOperationFlag]string{}
			ctx = context.WithValue(ctx, ckFilterQueries{}, existing)
		}
		for flag, p := range policies {
			if len(p) != 0 {
				existing[flag] = p
			}
		}
		tx.Statement.Context = ctx
		return tx
	}
}

/********************
	Helpers
 ********************/

type ckFilterMode struct{}
type ckFilterQueries struct{}
type ckFilterExtraInput struct{}

func shouldSkip(ctx context.Context, flag DBOperationFlag, fallback policyMode) bool {
	if ctx == nil {
		return defaultPolicyMode.hasFlags(flag)
	}
	switch v := ctx.Value(ckFilterMode{}).(type) {
	case policyMode:
		return !v.hasFlags(flag)
	default:
		return !fallback.hasFlags(flag)
	}
}

func resolveQuery(ctx context.Context, flag DBOperationFlag, isPartial bool, meta *Metadata) string {
	// ad-hoc info
	if queries, ok := ctx.Value(ckFilterQueries{}).(map[DBOperationFlag]string); ok {
		if q, ok := queries[flag]; ok && len(q) != 0 {
			return finalizeQuery(q, isPartial)
		}
	}

	// declarative info
	pkg := meta.OPAPackage
	var policy string
	if p, ok := meta.Policies[flag]; ok && p != TagValueIgnore {
		policy = p
	}

	// fallbacks
	switch {
	case len(pkg) == 0 && len(policy) == 0:
		// everything default
		return ""
	case len(pkg) == 0:
		pkg = fmt.Sprintf("%s.%s", opa.PackagePrefixResource, meta.ResType)
	}
	return finalizeQuery(fmt.Sprintf("data.%s.%s", pkg, policy), isPartial)
}

func finalizeQuery(query string, isPartial bool) string {
	if isPartial {
		query = strings.ReplaceAll(query, "/", ".")
		if !strings.HasPrefix(query, "data.") {
			query = "data." + query
		}
	} else {
		query = strings.ReplaceAll(query, ".", "/")
		query = strings.TrimPrefix(query, "data/")
	}
	return query
}
