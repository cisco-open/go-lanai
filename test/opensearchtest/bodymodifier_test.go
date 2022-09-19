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
	modifier := IgnoreGJSONPaths(t, []string{"name.last"})
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

// TestBodyModifierController tests that we can Append and Clear the modifiers using
// the controller
func TestBodyModifierController(t *testing.T) {
	var modifierController MatcherBodyModifierController
	var options RecordOption
	options.MatchBodyModifiers = modifierController.Modifier()

	for _, _ = range *options.MatchBodyModifiers {
		t.Errorf("Expected no MatchBodyModifiers to be in options")
	}
	modifierController.Append(func(i *[]byte) { /* no op*/ })
	numberOfModifiers := 0
	for _, o := range *options.MatchBodyModifiers {
		numberOfModifiers++
		if o == nil {
			t.Errorf("o should not be nil right now")
		}
	}
	if numberOfModifiers != 1 {
		t.Errorf("expected there to be exactly 1 modifier, not: %v", numberOfModifiers)
	}
	modifierController.Clear()
	for _, _ = range *options.MatchBodyModifiers {
		t.Errorf("Expected no MatchBodyModifiers to be in options")
	}
	modifierController.Append(func(i *[]byte) {
		// no op
	})
	numberOfModifiers = 0
	for _, o := range *options.MatchBodyModifiers {
		numberOfModifiers++
		if o == nil {
			t.Errorf("o should not be nil right now")
		}
	}
	if numberOfModifiers != 1 {
		t.Errorf("expected there to be exactly 1 modifier, not: %v", numberOfModifiers)
	}
}
