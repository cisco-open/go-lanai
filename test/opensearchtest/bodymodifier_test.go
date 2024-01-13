package opensearchtest

import (
	"encoding/json"
	"testing"
)

var TestJSON = `
{
  "name": {"first": "Tom", "last": "Anderson"},
  "time": "2022-09-15T14:59:32.929764Z"
}
`

func TestSetPath(t *testing.T) {
	modifier := IgnoreGJSONPaths(t, "name.last")
	byteTestJSON := []byte(TestJSON)
	modifier(&byteTestJSON)
	var testJSONStruct = struct {
		Name struct {
			First string `json:"first"`
			Last  string `json:"last"`
		} `json:"name"`
	}{}
	err := json.Unmarshal(byteTestJSON, &testJSONStruct)
	if err != nil {
		t.Fatalf("unable to unmarshal outputJSON: %v", err)
	}
	if testJSONStruct.Name.Last != "" {
		t.Errorf("expected last name to be empty, got: %v", testJSONStruct.Name.Last)
	}

}

// TestBodyModifiers tests that we can Append and Clear the modifiers
func TestBodyModifiers(t *testing.T) {
	modifiers := &MatcherBodyModifiers{}

	for _, _ = range modifiers.Modifier() {
		t.Errorf("Expected no MatchBodyModifiers to be in options")
	}
	modifiers.Append(func(i *[]byte) { /* no op*/ })
	numberOfModifiers := 0
	for _, o := range modifiers.Modifier() {
		numberOfModifiers++
		if o == nil {
			t.Errorf("o should not be nil right now")
		}
	}
	if numberOfModifiers != 1 {
		t.Errorf("expected there to be exactly 1 modifier, not: %v", numberOfModifiers)
	}
	modifiers.Clear()
	for _, _ = range modifiers.Modifier() {
		t.Errorf("Expected no MatchBodyModifiers to be in options")
	}
	modifiers.Append(func(i *[]byte) {
		// no op
	})
	numberOfModifiers = 0
	for _, o := range modifiers.Modifier() {
		numberOfModifiers++
		if o == nil {
			t.Errorf("o should not be nil right now")
		}
	}
	if numberOfModifiers != 1 {
		t.Errorf("expected there to be exactly 1 modifier, not: %v", numberOfModifiers)
	}
}
