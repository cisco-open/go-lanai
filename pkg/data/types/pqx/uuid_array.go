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
	"database/sql/driver"
	"fmt"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type UUIDArray []uuid.UUID

// Value implements driver.Valuer
func (a UUIDArray) Value() (driver.Value, error) {
	if a == nil {
		return nil, nil
	}
	return pq.StringArray(a.Strings()).Value()
}

// Scan implements sql.Scanner
func (a *UUIDArray) Scan(src interface{}) error {
	if a == nil {
		return nil
	}

	strArray := &pq.StringArray{}
	if e := strArray.Scan(src); e != nil {
		return e
	}
	uuids := make(UUIDArray, len(*strArray))
	for i, v := range *strArray {
		var e error
		if uuids[i], e = uuid.Parse(v); e != nil {
			return fmt.Errorf("pq: cannot convert %T to UUIDArray - %v", src, e)
		}
	}
	*a = uuids
	return nil
}

func (a UUIDArray) Strings() []string {
	strArray := make(pq.StringArray, len(a))
	for i, v := range a{
		strArray[i] = v.String()
	}
	return strArray
}
