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

package pqx

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/utils"
	"database/sql/driver"
	"fmt"
	"time"
)

// Duration is also an alias of time.Duration
type Duration utils.Duration

// Value implements driver.Valuer
func (d Duration) Value() (driver.Value, error) {
	return time.Duration(d).String(), nil
}

// Scan implements sql.Scanner
func (d *Duration) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		*d = Duration(utils.ParseDuration(string(src)))
	case string:
		*d = Duration(utils.ParseDuration(src))
	case int, int8, int16, int32, int64:
		// TODO review how convert numbers to Duration
		*d = Duration(src.(int64))
	case nil:
		return nil
	default:
		return data.NewDataError(data.ErrorCodeOrmMapping,
			fmt.Sprintf("pqx: unable to convert data type %T to Duration", src))
	}
	return nil
}

// MarshalText implements encoding.TextMarshaler
func (d Duration) MarshalText() (text []byte, err error) {
	return utils.Duration(d).MarshalText()
}

// UnmarshalText implements encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	return (*utils.Duration)(d).UnmarshalText(text)
}