package opadata

import (
	"context"
	"github.com/open-policy-agent/opa/sdk"
	"gorm.io/gorm"
)

type ContextAwarePartialQueryMapper interface {
	sdk.PartialQueryMapper
	WithContext(ctx context.Context) ContextAwarePartialQueryMapper
	Context() context.Context
}

type ckFilteringMode struct{}

const (
	FilteringFlagWriteValueCheck FilteringFlag = 1 << iota
	FilteringFlagReadFiltering
	FilteringFlagWriteFiltering
	FilteringFlagDeleteFiltering
)

// FilteringFlag bitwise Flag of tenancy flag mode
type FilteringFlag uint

const (
	filteringModeDefault = filteringMode(FilteringFlagReadFiltering | FilteringFlagWriteFiltering | FilteringFlagWriteValueCheck)
)

// filteringMode enum of tenancyCheckMode
type filteringMode uint

func (m filteringMode) hasFlags(flags ...FilteringFlag) bool {
	for _, flag := range flags {
		if m&filteringMode(flag) == 0 {
			return false
		}
	}
	return true
}

// SkipPolicyFiltering is used as a scope for gorm.DB to skip tenancy check
// e.g. db.WithContext(ctx).Scopes(SkipPolicyFiltering()).Find(...)
// Note using this scope without context would panic
func SkipPolicyFiltering() func(*gorm.DB) *gorm.DB {
	return FilterByPolicy(0)
}

// FilterByPolicy is used as a scope for gorm.DB to override tenancy check
// e.g. db.WithContext(ctx).Scopes(FilterByPolicy()).Find(...)
// Note using this scope without context would panic
func FilterByPolicy(flags ...FilteringFlag) func(*gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if tx.Statement.Context == nil {
			panic("SkipPolicyFiltering used without context")
		}
		var mode filteringMode
		for _, flag := range flags {
			mode = mode | filteringMode(flag)
		}
		ctx := context.WithValue(tx.Statement.Context, ckFilteringMode{}, mode)
		tx.Statement.Context = ctx
		return tx
	}
}
