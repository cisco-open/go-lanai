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
    "encoding/json"
    "fmt"
    "github.com/cisco-open/go-lanai/pkg/data"
)

// JsonbScan helps models to implement sql.Scanner
func JsonbScan(src interface{}, v interface{}) error {
	var d []byte
	switch src.(type) {
	case string:
		d = []byte(src.(string))
	case []byte:
		d = src.([]byte)
	case nil:
		return nil
	default:
		msg := fmt.Sprintf("unable to scan %T as JSONB format", src)
		return data.NewDataError(data.ErrorCodeOrmMapping, msg)
	}
	if e := json.Unmarshal(d, v); e != nil {
		return data.NewDataError(data.ErrorCodeOrmMapping, fmt.Sprintf("unable to scan JSONB into %T: %v", v, e), e)
	}
	return nil
}

// JsonbValue helps models to implement driver.Valuer
func JsonbValue(v interface{}) (driver.Value, error) {
	if v == nil {
		return nil, nil
	}

	d, e := json.Marshal(v)
	if e != nil {
		return nil, data.NewDataError(data.ErrorCodeInvalidSQL, fmt.Sprintf("unable to convert %T to JSONB: %v", v, e), e)
	}

	return string(d), nil
}

type JsonbMap map[string]interface{}

func (m JsonbMap) Value() (driver.Value, error) {
	return JsonbValue(m)
}

func (m *JsonbMap) Scan(src interface{}) error {
	return JsonbScan(src, m)
}

type JsonbStringMap map[string]string

func (m JsonbStringMap) Value() (driver.Value, error) {
	return JsonbValue(m)
}

func (m *JsonbStringMap) Scan(src interface{}) error {
	return JsonbScan(src, m)
}