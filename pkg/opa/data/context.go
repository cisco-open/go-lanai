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

package opadata

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"fmt"
	"strings"
)

/*****************
	Constants
 *****************/

const (
	DefaultQueryTemplate = `allow_%s`
	DefaultPartialQueryTemplate = `filter_%s`
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

// OPATag supported key-value pairs in `opa` tag.
// `opa` tag is in format of `opa:"<key>:<value>, [<more_keys>:<more_values>, ...]".
// Unless specified, each key-value pair only takes effect on either "to-be-filtered-by" model fields (Model Fields)
// or FilteredModel (regardless if embedded or as a field), but not both.
type OPATag struct {
	// InputField Required on "to-be-filtered-by" model fields. Specify mappings between model field and OPA input fields.
	// e.g. `opa:"field:myProperty"` translate to `input.resource.myProperty` in OPA input
	InputField string

	// ResType Required on FilteredModel. This value contributes to both OPA query and OPA input:
	// - ResType is set to OPA input as `input.resource.type`
	// - Unless OPAPackage or Policies is specified, ResType is also part of OPA query:
	//   "data.resource.{{RestType}}.<filter|allow>_{{DBOperationFlag}}"
	ResType string

	// OPAPackage Optional on FilteredModel. Used to overwrite default OPA query.
	// Resulting query is "data.{{OPAPackage}}.<filter|allow>_{{DBOperationFlag}}"
	// e.g. `opa:"type:my_res, package:my.res" -> the OPA query is "data.my.res.filter_{{DBOperationFlag}}"
	OPAPackage string

	// Policies Optional on FilteredModel. Fine control of OPA queries for each type of DB operation.
	// - If set to "-", the corresponding DB operation is disabled for data-filtering.
	// 	 e.g. `opa:"type:my_res, read:-"` disables OPA data filtering for read operations (SELECT statements)
	// - If set to any other non-empty string, it's used to construct OPA query
	// 	 e.g. `opa:"type:my_res, read:my_custom_filter"` -> OPA query "data.resource.my_res.my_custom_filter" is used for read operations.
	Policies map[DBOperationFlag]string

	// mode bitwise flags for enabled/disabled DB operations
	mode     policyMode
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
