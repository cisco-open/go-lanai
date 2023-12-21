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

package utils

import (
	"errors"
	"time"
)

const (
	ISO8601Seconds      = "2006-01-02T15:04:05Z07:00" //time.RFC3339
	ISO8601Milliseconds = "2006-01-02T15:04:05.000Z07:00"
)

var (
	MaxTime = time.Unix(1<<63-1, 0).UTC()
)

func ParseTimeISO8601(v string) time.Time {
	parsed, err := time.Parse(ISO8601Seconds, v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func ParseTime(layout, v string) time.Time {
	parsed, err := time.Parse(layout, v)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

func ParseDuration(v string) time.Duration {
	parsed, err := time.ParseDuration(v)
	if err != nil {
		return time.Duration(0)
	}
	return parsed
}

type Duration time.Duration

// encoding.TextMarshaler
func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(time.Duration(d).String()), nil
}

// encoding.TextUnmarshaler
func (d *Duration) UnmarshalText(text []byte) error {
	if d == nil {
		return errors.New("duration pointer is nil")
	}

	parsed, e := time.ParseDuration(string(text))
	if e != nil {
		return e
	}

	*d = Duration(parsed)
	return nil
}
