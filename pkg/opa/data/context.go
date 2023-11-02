package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
	"strings"
)

/*****************
	Errors
 *****************/

var (
	ErrQueryTranslation = opa.NewError(`generic query translation error`)
	ErrUnsupportedUsage = opa.NewError(`generic unsupported usage error`)
)

/*****************
	Tag
 *****************/

const (
	TagOPA              = `opa`
	TagDelimiter        = `,`
	TagAssignment       = `:`
	TagValueIgnore      = "-"
	TagKeyInputField    = `field`
	TagKeyInputFieldAlt = `input`
	TagKeyResourceType  = `type`
	TagKeyOPAPackage    = `package`
)

type OPATag struct {
	// InputField applies to "to-be-filtered-by" model fields.
	InputField string
	ResType    string
	OPAPackage string
	Policies   map[DBOperationFlag]string
	mode       policyMode
}

func (t *OPATag) UnmarshalText(data []byte) error {
	// setup default
	*t = OPATag{
		mode: defaultPolicyMode,
	}
	// parse kv pairs
	terms := strings.Split(string(data), TagDelimiter)
	for _, term := range terms {
		term = strings.TrimSpace(term)
		if len(term) == 0 {
			continue
		}
		kv := strings.SplitN(term, TagAssignment, 2)
		var v string
		switch len(kv) {
		case 1:
			v = `true`
		case 2:
			v = strings.TrimSpace(kv[1])
		default:
			return fmt.Errorf(`invalid "opa" tag format, expect "key:model", but got "%s"`, term)
		}
		k := strings.TrimSpace(kv[0])
		switch k {
		case TagKeyInputField, TagKeyInputFieldAlt:
			t.InputField = v
		case TagKeyResourceType:
			t.ResType = v
		case TagKeyOPAPackage:
			t.OPAPackage = v
		default:
			if e := t.parsePolicy(kv); e == nil {
				continue
			}
			return ErrUnsupportedUsage.WithMessage(`invalid "opa" tag, unrecognized key "%s"`, k)
		}
	}
	return nil
}

func (t *OPATag) parsePolicy(kv []string) error {
	if len(kv) != 2 {
		return fmt.Errorf(`invalid policy, expect <mode>%s<policy_name>`, TagAssignment)
	}
	var flag DBOperationFlag
	if e := flag.UnmarshalText([]byte(strings.TrimSpace(kv[0]))); e != nil {
		return e
	}
	if t.Policies == nil {
		t.Policies = map[DBOperationFlag]string{}
	}
	t.Policies[flag] = strings.TrimSpace(kv[1])
	if kv[1] == TagValueIgnore {
		t.mode = t.mode & ^policyMode(flag)
	} else {
		t.mode = t.mode | policyMode(flag)
	}
	return nil
}

// Queries normalized queries for OPA.
// By default, queries are
func (t *OPATag) Queries() map[DBOperationFlag]string {
	//TODO
	return t.Policies
}

/********************
	Flags and Mode
 ********************/

const (
	DBOperationFlagCreate DBOperationFlag = 1 << iota
	DBOperationFlagRead
	DBOperationFlagUpdate
	DBOperationFlagDelete
)

const (
	dbOpTextCreate = `create`
	dbOpTextRead   = `read`
	dbOpTextUpdate = `update`
	dbOpTextDelete = `delete`
)

// DBOperationFlag bitwise Flag of tenancy flag mode
type DBOperationFlag uint

func (f DBOperationFlag) MarshalText() ([]byte, error) {
	switch f {
	case DBOperationFlagCreate:
		return []byte(dbOpTextCreate), nil
	case DBOperationFlagRead:
		return []byte(dbOpTextRead), nil
	case DBOperationFlagUpdate:
		return []byte(dbOpTextUpdate), nil
	case DBOperationFlagDelete:
		return []byte(dbOpTextDelete), nil
	}
	return []byte{}, nil
}

func (f *DBOperationFlag) UnmarshalText(data []byte) error {

	switch v := string(data); v {
	case dbOpTextCreate:
		*f = DBOperationFlagCreate
	case dbOpTextRead:
		*f = DBOperationFlagRead
	case dbOpTextUpdate:
		*f = DBOperationFlagUpdate
	case dbOpTextDelete:
		*f = DBOperationFlagDelete
	default:
		return fmt.Errorf("unrecognized DB operation flag '%s'", string(data))
	}
	return nil
}

const (
	defaultPolicyMode = policyMode(DBOperationFlagCreate | DBOperationFlagRead | DBOperationFlagUpdate | DBOperationFlagDelete)
)

// policyMode enum of policyMode
type policyMode uint

//goland:noinspection GoMixedReceiverTypes
func (m policyMode) hasFlags(flags ...DBOperationFlag) bool {
	for _, flag := range flags {
		if m&policyMode(flag) == 0 {
			return false
		}
	}
	return true
}
