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

package constraints

import (
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/data/types/pqx"
	"cto-github.cisco.com/NFV-BU/go-lanai/pkg/opa"
	"database/sql/driver"
	"github.com/google/uuid"
)

const (
	SharedPermissionRead = SharedPermission(opa.OpRead)
	SharedPermissionWrite = SharedPermission(opa.OpWrite)
	SharedPermissionDelete  = SharedPermission(opa.OpDelete)
)

type SharedPermission opa.ResourceOperation

// Sharing is a Model type that stores mapping between user IDs and a list of allowed permissions as JSONB map
// This type works with OPA sharing policy
type Sharing map[uuid.UUID][]SharedPermission

// Value implements driver.Valuer
func (s Sharing) Value() (driver.Value, error) {
	return pqx.JsonbValue(s)
}

// Scan implements sql.Scanner
func (s *Sharing) Scan(src interface{}) error {
	return pqx.JsonbScan(src, s)
}

func (s Sharing) GormDataType() string {
	return "jsonb"
}

func (s Sharing) Share(userID uuid.UUID, perms ...SharedPermission) {
	if len(perms) == 0 {
		delete(s, userID)
	} else {
		s[userID] = perms
	}
}
