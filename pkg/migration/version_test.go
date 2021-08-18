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
