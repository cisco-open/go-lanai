package utils

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