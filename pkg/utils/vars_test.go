package utils

import (
	"reflect"
	"testing"
)

func TestFromPtrWithString(t *testing.T) {
	type args struct {
		value *string
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "Resolve a valid string value",
			args: args{
				value: ToPtr("hello"),
			},
			want: "hello",
		},
		{
			name: "resolve an empty string from empty string value",
			args: args{
				value: ToPtr(""),
			},
			want: "",
		},
		{
			name: "get empty string from nil value",
			args: args{
				value: nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromPtr(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromPtrWithBool(t *testing.T) {
	type args struct {
		value *bool
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "resolve true from valid true pointer",
			args: args{
				value: ToPtr(true),
			},
			want: true,
		},
		{
			name: "resolve false from valid false pointer",
			args: args{
				value: ToPtr(false),
			},
			want: false,
		},
		{
			name: "resolve false from nil pointer",
			args: args{
				value: nil,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromPtr(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFromPtrWithFloat64(t *testing.T) {
	type args struct {
		value *float64
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "resolve float value from valid pointer",
			args: args{
				value: ToPtr(1.234),
			},
			want: 1.234,
		},
		{
			name: "resolve 0.0 from valid pointer",
			args: args{
				value: ToPtr(0.0),
			},
			want: 0.0,
		},
		{
			name: "resolve 0.0 from nil pointer",
			args: args{
				value: nil,
			},
			want: 0.0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromPtr(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestFromPtrWithUnderlyingTypeString tests that types that have an underlying
// type of any of the primitives still work. For example
//  type String string
//  type Integer int
// The above types should still work
func TestFromPtrWithUnderlyingTypeString(t *testing.T) {
	type String string
	type args struct {
		value *String
	}
	tests := []struct {
		name string
		args args
		want any
	}{
		{
			name: "Resolve a valid string value",
			args: args{
				value: ToPtr(String("hello")),
			},
			want: String("hello"),
		},
		{
			name: "resolve an empty string from empty string value",
			args: args{
				value: ToPtr(String("")),
			},
			want: String(""),
		},
		{
			name: "get empty string from nil value",
			args: args{
				value: nil,
			},
			want: String(""),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FromPtr(tt.args.value); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromPtr() = %v, want %v", got, tt.want)
			}
		})
	}
}
