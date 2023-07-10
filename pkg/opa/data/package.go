package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"gorm.io/gorm"
	"reflect"
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
// Note using this scope without context would panic
func SkipPolicyFiltering() func(*gorm.DB) *gorm.DB {
	return FilterByPolicy(0)
}

// FilterByPolicy is used as a scope for gorm.DB to override policy-based data filtering
// e.g. db.WithContext(ctx).Scopes(FilterByPolicy()).Find(...)
// Note using this scope without context would panic
func FilterByPolicy(flags ...DBOperationFlag) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("FilterByPolicy scope is used without context")
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

/********************
	Helpers
 ********************/

type ckFilterMode struct{}

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
