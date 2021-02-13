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