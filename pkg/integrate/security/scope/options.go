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

package scope

func WithUsername(username string) Options {
	return func(s *Scope) {
		s.username = username
		s.userId = ""
	}
}

func WithUserId(userId string) Options {
	return func(s *Scope) {
		s.username = ""
		s.userId = userId
	}
}

func WithTenantId(tenantId string) Options {
	return func(s *Scope) {
		s.tenantExternalId = ""
		s.tenantId = tenantId
	}
}

func WithTenantExternalId(tenantExternalId string) Options {
	return func(s *Scope) {
		s.tenantExternalId = tenantExternalId
		s.tenantId = ""
	}
}

func UseSystemAccount() Options {
	return func(s *Scope) {
		s.useSysAcct = true
	}
}
