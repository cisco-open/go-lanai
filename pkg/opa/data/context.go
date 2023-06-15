package opadata

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/log"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
	"gorm.io/gorm"
	"strings"
)

var logger = log.New("OPA.Data")

/*****************
	Errors
 *****************/

var (
	QueryTranslationError = opa.NewError(`generic query translation error`)
)

const (
	TagOPA              = `opa`
	TagDelimiter        = `;`
	TagAssignment       = `:`
	TagKeyInputField    = `field`
	TagKeyInputFieldAlt = `input`
	TagKeyResourceType  = `type`
	TagKeyPolicy        = `policy`
)

/*****************
	Tag
 *****************/

type opaTag struct {
	InputField string
	ResType    string
	Policy     string
}

func (t *opaTag) UnmarshalText(data []byte) error {
	terms := strings.Split(string(data), TagDelimiter)
	for _, term := range terms {
		kv := strings.SplitN(term, TagAssignment, 2)
		var v string
		switch len(kv) {
		case 1:
			v = `true`
		case 2:
			v = strings.TrimSpace(kv[1])
		default:
			return fmt.Errorf(`invalid "opa" tag format, expect "key:value", but got "%s"`, term)
		}
		k := kv[0]
		switch k {
		case TagKeyInputField, TagKeyInputFieldAlt:
			t.InputField = v
		case TagKeyResourceType:
			t.ResType = v
		case TagKeyPolicy:
			t.Policy = v
		default:
			return fmt.Errorf(`invalid "opa" tag, unrecognized key "%s"`, k)
		}
	}
	return nil
}

/********************
	Flags and Mode
 ********************/

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

/********************
	GORM Scopes
 ********************/

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
