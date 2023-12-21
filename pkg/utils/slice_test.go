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
	"reflect"
	"testing"
)

func TestCommaSeparatedSlice_UnmarshalText(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		s        CommaSeparatedSlice
		args     args
		expected CommaSeparatedSlice
		wantErr  bool
	}{
		{
			name: "UnmarshalText should separate text into comma seperated slices",
			s:    CommaSeparatedSlice{},
			args: args{data: []byte("hello, world")},
			expected: CommaSeparatedSlice{
				"hello",
				"world",
			},
			wantErr: false,
		},
		{
			name: "UnmarshalText should trim leading/trailing spaces from input",
			s:    CommaSeparatedSlice{},
			args: args{data: []byte("trailing , leading, leading and trailing ")},
			expected: CommaSeparatedSlice{
				"trailing",
				"leading",
				"leading and trailing",
			},
			wantErr: false,
		},
		{
			name:     "UnmarshalText should return empty slice if provided empty string",
			s:        CommaSeparatedSlice{},
			args:     args{data: []byte("")},
			expected: CommaSeparatedSlice{},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.s.UnmarshalText(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalText() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.s, tt.expected) {
				t.Errorf("Failed: expected %v, got %v", tt.expected, tt.s)
			}
		})
	}
}

func TestRemoveIntSlices(t *testing.T) {
	type args struct {
		slice []int
		i     int
	}
	tests := []struct {
		name        string
		args        args
		want        []int
		expectPanic bool
	}{
		{
			name: "simple int slice",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				i:     3,
			},
			want: []int{1, 2, 3, 5},
		},
		{
			name: "remove 0th index",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				i:     0,
			},
			want: []int{5, 2, 3, 4},
		},
		{
			name: "remove last index",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				i:     4,
			},
			want: []int{1, 2, 3, 4},
		},
		{
			name: "remove index out of bounds +",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				i:     5,
			},
			want:        []int{1, 2, 3, 4, 5},
			expectPanic: true,
		},
		{
			name: "remove index out of bounds -",
			args: args{
				slice: []int{1, 2, 3, 4, 5},
				i:     -5,
			},
			want:        []int{1, 2, 3, 4, 5},
			expectPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected code to panic but the test did not")
					}
				}()
			}
			if got := Remove(tt.args.slice, tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveStructSlices(t *testing.T) {
	type placeholder struct {
		ID int
	}
	type args struct {
		slice []placeholder
		i     int
	}
	tests := []struct {
		name        string
		args        args
		want        []placeholder
		expectPanic bool
	}{
		{
			name: "remove middle",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     1,
			},
			want: []placeholder{{ID: 1}, {ID: 3}},
		},
		{
			name: "remove 0th index",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     0,
			},
			want: []placeholder{{ID: 3}, {ID: 2}},
		},
		{
			name: "remove last index",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     2,
			},
			want: []placeholder{{ID: 1}, {ID: 2}},
		},
		{
			name: "remove index out of bounds +",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     5,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
		{
			name: "remove index out of bounds -",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     -1,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
		{
			name: "remove index out of bounds with length len(slice)",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     3,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected code to panic but the test did not")
					}
				}()
			}
			if got := Remove(tt.args.slice, tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveStableStructSlices(t *testing.T) {
	type placeholder struct {
		ID int
	}
	type args struct {
		slice []placeholder
		i     int
	}
	tests := []struct {
		name        string
		args        args
		want        []placeholder
		expectPanic bool
	}{
		{
			name: "remove middle",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     1,
			},
			want: []placeholder{{ID: 1}, {ID: 3}},
		},
		{
			name: "remove 0th index",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     0,
			},
			want: []placeholder{{ID: 2}, {ID: 3}},
		},
		{
			name: "remove last index",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     2,
			},
			want: []placeholder{{ID: 1}, {ID: 2}},
		},
		{
			name: "remove middle - larger slice",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 5}, {ID: 8}, {ID: 9}},
				i:     3,
			},
			want: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}, {ID: 8}, {ID: 9}},
		},
		{
			name: "remove index out of bounds +",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     5,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
		{
			name: "remove index out of bounds -",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     -1,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
		{
			name: "remove index out of bounds with length len(slice)",
			args: args{
				slice: []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
				i:     3,
			},
			want:        []placeholder{{ID: 1}, {ID: 2}, {ID: 3}},
			expectPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("expected code to panic but the test did not")
					}
				}()
			}
			if got := RemoveStable(tt.args.slice, tt.args.i); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Remove() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReverse(t *testing.T) {
	type args struct {
		input []any
	}
	tests := []struct {
		name     string
		args     args
		expected []any
	}{
		{
			name: "simple ints",
			args: args{
				input: []any{1, 3, 2, 4},
			},
			expected: []any{4, 2, 3, 1},
		},
		{
			name: "simple strings",
			args: args{
				input: []any{"5", "4", "1", "3"},
			},
			expected: []any{"3", "1", "4", "5"},
		},
		{
			name: "int and string mix",
			args: args{
				input: []any{5, "4", "1", "3"},
			},
			expected: []any{"3", "1", "4", 5},
		},
		{
			name: "string and struct mix",
			args: args{
				input: []any{struct{ s string }{s: "hello"}, "4", "1", "3"},
			},
			expected: []any{"3", "1", "4", struct{ s string }{s: "hello"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Reverse(tt.args.input)
			for i, value := range tt.expected {
				if value != tt.args.input[i] {
					t.Fatalf("Reverse() = %v, want %v", tt.args, tt.expected)
				}
			}
		})
	}
}
