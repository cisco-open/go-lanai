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

/*****************
	Tag
 *****************/

const (
	TagOPA              = `opa`
	TagDelimiter        = `;`
	TagAssignment       = `:`
	TagKeyInputField    = `field`
	TagKeyInputFieldAlt = `input`
	TagKeyResourceType  = `type`
	TagKeyPolicy        = `policy`
	TagKeyMode          = `mode`
)

type opaTag struct {
	InputField string
	ResType    string
	Policy     string
	Mode       policyMode
}

func (t *opaTag) UnmarshalText(data []byte) error {
	// setup default
	*t = opaTag{
		Mode: defaultPolicyMode,
	}
	// parse kv pairs
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
		case TagKeyMode:
			if e := t.Mode.UnmarshalText([]byte(v)); e != nil {
				return fmt.Errorf(`invalid "opa" tag, unrecognized mode "%s"`, v)
			}
		default:
			return fmt.Errorf(`invalid "opa" tag, unrecognized key "%s"`, k)
		}
	}
	return nil
}

/********************
	Flags and Mode
 ********************/

const (
	PolicyFlagCreate PolicyFlag = 1 << iota
	PolicyFlagRead
	PolicyFlagUpdate
	PolicyFlagDelete
)

// PolicyFlag bitwise Flag of tenancy flag mode
type PolicyFlag uint

const (
	defaultPolicyMode = policyMode(PolicyFlagCreate | PolicyFlagRead | PolicyFlagUpdate | PolicyFlagDelete)
)

// policyMode enum of policyMode
type policyMode uint

//goland:noinspection GoMixedReceiverTypes
func (m policyMode) hasFlags(flags ...PolicyFlag) bool {
	for _, flag := range flags {
		if m&policyMode(flag) == 0 {
			return false
		}
	}
	return true
}

//goland:noinspection GoMixedReceiverTypes
func (m *policyMode) UnmarshalText(data []byte) error {
	//TODO
	*m = defaultPolicyMode
	return nil
}

/********************
	GORM Scopes
 ********************/

type ckFilterMode struct{}

// SkipPolicyFiltering is used as a scope for gorm.DB to skip tenancy check
// e.g. db.WithContext(ctx).Scopes(SkipPolicyFiltering()).Find(...)
// Note using this scope without context would panic
func SkipPolicyFiltering() func(*gorm.DB) *gorm.DB {
	return FilterByPolicy(0)
}

// FilterByPolicy is used as a scope for gorm.DB to override tenancy check
// e.g. db.WithContext(ctx).Scopes(FilterByPolicy()).Find(...)
// Note using this scope without context would panic
func FilterByPolicy(flags ...PolicyFlag) func(*gorm.DB) *gorm.DB {
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

func shouldSkip(ctx context.Context, flag PolicyFlag, fallback policyMode) bool {
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
