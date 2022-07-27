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
