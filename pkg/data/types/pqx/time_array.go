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
    "bytes"
    "database/sql/driver"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
    "github.com/lib/pq"
    "time"
)

// TimeArray register driver.Valuer & sql.Scanner
type TimeArray []time.Time

// Value implements driver.Valuer
func (a TimeArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}

	n := len(a)
	if n <= 0 {
		return "{}", nil
	}

	// There will be at least two curly brackets, 2*N bytes of quotes,
	// and N-1 bytes of delimiters.
	b := make([]byte, 1, 1+3*n)
	b[0] = '{'

	b = appendArrayQuotedBytes(b, pq.FormatTimestamp(a[0]))
	for i := 1; i < n; i++ {
		b = append(b, ',')
		b = appendArrayQuotedBytes(b, pq.FormatTimestamp(a[i]))
	}

	return string(append(b, '}')), nil
}

// Scan implements sql.Scanner
func (a *TimeArray) Scan(src interface{}) error {
	switch src := src.(type) {
	case []byte:
		return a.scanBytes(src)
	case string:
		return a.scanBytes([]byte(src))
	case nil:
		*a = nil
		return nil
	}

	return fmt.Errorf("pqx: cannot convert %T to TimeArray", src)
}

func (a *TimeArray) scanBytes(src []byte) error {
	var strs pq.StringArray
	sPtr := &strs
	if e := sPtr.Scan(src); e != nil {
		return data.NewDataError(data.ErrorCodeOrmMapping, e)
	}

	elems := make(TimeArray, len(strs))
	for i, s := range strs {
		t, e := pq.ParseTimestamp(time.UTC, s)
		if e != nil {
			return data.NewDataError(data.ErrorCodeOrmMapping,
				fmt.Sprintf("pqx: parsing array at idx %d: %v", i, e.Error()), e)
		}
		elems[i] = t
	}
	*a = elems
	return nil
}

func appendArrayQuotedBytes(b, v []byte) []byte {
	b = append(b, '"')
	for {
		i := bytes.IndexAny(v, `"\`)
		if i < 0 {
			b = append(b, v...)
			break
		}
		if i > 0 {
			b = append(b, v[:i]...)
		}
		b = append(b, '\\', v[i])
		v = v[i+1:]
	}
	return append(b, '"')
}