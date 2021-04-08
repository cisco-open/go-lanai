package utils

import "github.com/google/uuid"

var (
	TRUE  = true
	FALSE = false
)

func BoolPtr(v bool) *bool {
	if v {
		return &TRUE
	} else {
		return &FALSE
	}
}

func IntPtr(v int) *int {
	return &v
}

func UIntPtr(v uint) *uint {
	return &v
}

func Float64Ptr(v float64) *float64 {
	return &v
}

func StringPtr(v string) *string {
	return &v
}

func UuidPtr(v uuid.UUID) *uuid.UUID {
	return &v
}
