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

package migration

import "testing"

func TestVersionComparison(t *testing.T) {
	v4001, _ := fromString("v4.0.0.1")
	v40010, _ := fromString("4.0.0.10")
	v4002, _ := fromString("4.0.0.2")

	if !v4001.Lt(v40010) {
		t.Errorf("%v should be less than %v", v4001, v40010)
	}

	if !v4002.Lt(v40010) {
		t.Errorf("%v should be less than %v", v4001, v40010)
	}
}
