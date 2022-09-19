package opensearchtest

import (
	"github.com/tidwall/sjson"
	"testing"
)

// MatcherBodyModifierController provides a way to control the MatcherBodyModifier that is
// passed into the MatchBody function.
type MatcherBodyModifierController struct {
	modifier []MatcherBodyModifier
}

// Modifier will return pointer to the slice of MatcherBodyModifier
func (m *MatcherBodyModifierController) Modifier() *[]MatcherBodyModifier {
	return &m.modifier
}

// Append can be used to append a new MatcherBodyModifier to the controller
func (m *MatcherBodyModifierController) Append(modifier MatcherBodyModifier) {
	m.modifier = append(m.modifier, modifier)
}

// Clear is used to clear all of the existing MatcherBodyModifier in the controller
func (m *MatcherBodyModifierController) Clear() {
	m.modifier = nil
}

// MatcherBodyModifier will modify the body of a request that goes to the MatchBody
// to remove things that might make matching difficult.
// Example being time parameters in queries, or randomly generated values.
// To see this in use, check out SubTestTimeBasedQuery in opensearch_test.go
type MatcherBodyModifier func(*[]byte)

// IgnoreGJSONPaths will ignore any of the fields that are defined by the gjsonPaths
// which follow the GJSON syntax.
// https://github.com/tidwall/gjson/blob/master/SYNTAX.md#gjson-path-syntax
func IgnoreGJSONPaths(t *testing.T, gjsonPaths []string) MatcherBodyModifier {
	return func(b *[]byte) {
		var err error
		for _, path := range gjsonPaths {
			*b, err = sjson.DeleteBytes(*b, path)
			if err != nil {
				t.Errorf("unable to delete bytes: %v", err)
			}
		}
	}
}
