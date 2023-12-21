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

package generator

import "testing"

func Test_filenameFromPath(t *testing.T) {
	type args struct {
		pathName string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Should convert an api path to a filename",
			args: args{pathName: "/my/api/v1/testpath/{scope}"},
			want: "testpath_scope.go",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filenameFromPath(tt.args.pathName); got != tt.want {
				t.Errorf("filenameFromPath() = %v, want %v", got, tt.want)
			}
		})
	}
}
