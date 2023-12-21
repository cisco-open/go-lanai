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

package errorutils

const (
	// reserved

	ReservedOffset			= 32
	ReservedMask			= ^int64(0) << ReservedOffset

	// error type bits

	ErrorTypeOffset = 24
	ErrorTypeMask   = ^int64(0) << ErrorTypeOffset

	// error sub type bits
	
	ErrorSubTypeOffset = 12
	ErrorSubTypeMask   = ^int64(0) << ErrorSubTypeOffset

	DefaultErrorCodeMask = ^int64(0)
)

