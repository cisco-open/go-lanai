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

package common

import (
	"fmt"
	"time"
)

const RedisNameSpace = "LANAI:SESSION" //This is to avoid confusion with records from other frameworks.
const SessionLastAccessedField = "lastAccessed"
const SessionIdleTimeoutDuration = "idleTimeout"
const SessionAbsTimeoutTime = "absTimeout"
const DefaultName = "SESSION"

type TimeoutSetting int

const (
	IdleTimeoutEnabled     TimeoutSetting = 1 << iota
	AbsoluteTimeoutEnabled TimeoutSetting = 1 << iota
)

func GetRedisSessionKey(name string, id string) string {
	return fmt.Sprintf("%s:%s:%s", RedisNameSpace, name, id)
}

func CalculateExpiration(setting TimeoutSetting, idleExpiration time.Time, absExpiration time.Time) (canExpire bool, expiration time.Time) {
	switch setting {
	case AbsoluteTimeoutEnabled:
		return true, absExpiration
	case IdleTimeoutEnabled:
		return true, idleExpiration
	case AbsoluteTimeoutEnabled | IdleTimeoutEnabled:
		//whichever is the earliest
		if idleExpiration.Before(absExpiration) {
			return true, idleExpiration
		} else {
			return true, absExpiration
		}
	default:
		return false, time.Time{}
	}
}