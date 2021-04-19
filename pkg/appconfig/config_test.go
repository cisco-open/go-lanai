package appconfig

import (
	"bytes"
	"encoding/gob"
	"fmt"
	. "github.com/onsi/gomega"
	"strings"
	"testing"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

func TestParsePlaceHolders(t *testing.T) {
	v := "${my.place.holder}"
	placeHolderKeys, isEmbedded, error := parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != false || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "1${my.place.holder}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "${my.place.holder}1"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 1 || strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}


	v = "${my.place.holder}${my.second.holder}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)

	if len(placeHolderKeys) != 2 || isEmbedded != true || error != nil {
		t.Errorf("%s not parsed correctly", v)
	}

	if strings.Compare(placeHolderKeys[0], "my.place.holder") != 0 {
		t.Errorf("%s not parsed correctly", v)
	}

	if strings.Compare(placeHolderKeys[1], "my.second.holder") != 0 {
		t.Errorf("%s not parsed correctly", v)
	}

	v = "${${my.second.holder}}"
	placeHolderKeys, isEmbedded, error = parsePlaceHolder(v)
	if error == nil {
		t.Errorf("%s not parsed correctly. Expected error", v)
	}
	fmt.Println(error)
}

func TestUnFlatten(t *testing.T) {
	flat := map[string]interface{} {
		"e.f.g[0]" : "f_g_0",
		"e.f.g[1]" : "f_g_1",
	}
	nested, err := UnFlatten(flat)

	if err != nil {
		t.Errorf("Couldn't un-flatten map")
	}

	e := nested["e"].(map[string]interface{})
	f := e["f"].(map[string]interface{})
	g := f["g"].([]interface{})

	if len(g) != 2 {
		t.Errorf("second level array not unflattened properly")
	}

	flat = map[string]interface{} {
		"e.f[0].g[0]" : "f_0_g_0",
	}

	nested, err = UnFlatten(flat)

	if err == nil {
		t.Errorf("expected error")
	}

	fmt.Println(err)
}

func TestGettingValue(t *testing.T) {
	nested := make(map[string]interface{})
	v := value(nested, "a.b.c")

	if v != nil {
		t.Errorf("should get nil value")
	}

	nested = map[string]interface{}{
		"a" : map[string]interface{} {
			"b":map[string]interface{} {
				"c":"c_value",
			},
		},
	}

	v = value(nested, "a.b.c")

	if v != "c_value" {
		t.Errorf("expected %s, got %s", "c_value", v)
	}
}

func TestGettingArrayValue(t *testing.T) {
	nested := map[string]interface{}{
		"a" : map[string]interface{}{
			"b":[]interface{}{
				map[string]interface{}{
					"c":[]interface{}{
						"c1",
					},
				},
			},
		},
	}
	v := value(nested, "a.b[1]")

	if v != nil {
		t.Errorf("should get nil value")
	}

	v = value(nested, "a.b[0].c[1]")

	if v != nil {
		t.Errorf("should get nil value")
	}

	v = value(nested, "a.b[0].c[0]")
	if v != "c1" {
		t.Errorf("expected %s, actual %s", "c1", v)
	}
}

func TestSettingValue(t *testing.T) {
	var values map[string]interface{}

	values = map[string]interface{}{}
	t.Run("EmptyMapTestNoIntermediateNodes", SetValueTest(values, "a.b.c", false, true))
	t.Run("EmptyMapTest", SetValueTest(values, "a.b.c", true, false))

	values = map[string]interface{}{
		"a": map[string]interface{}{},
	}
	t.Run("PartialMapTestNoIntermediateNodes", SetValueTest(values, "a.b.c", false, true))
	t.Run("PartialMapTest", SetValueTest(values, "a.b.c", true, false))

	values = map[string]interface{}{
		"a": map[string]interface{}{
			"b": map[string]interface{}{
				"c": "another value",
			},
		},
	}
	t.Run("FullMapTestNoIntermediateNodes", SetValueTest(values, "a.b.c", false, false))
	t.Run("FullMapTest", SetValueTest(values, "a.b.c", true, false))
}

func TestSettingValueWithSlicePath(t *testing.T) {
	var values map[string]interface{}

	values = map[string]interface{}{}
	t.Run("EmptyMapTestNoIntermediateNodes", SetValueTest(values, "a.b[0].c[0]", false, true))
	t.Run("EmptyMapTest", SetValueTest(values, "a.b[0].c[0]", true, false))
	t.Run("EmptyMapTestNonFirstIntermediateNodes", SetValueTest(values, "a.b[1].c[1]", true, false))

	values = map[string]interface{}{
		"a": map[string]interface{}{
			"b": []interface{}{
				map[string]interface{}{},
			},
		},
	}
	t.Run("PartialMapTestNoIntermediateNodes", SetValueTest(values, "a.b[0].c[0]", false, true))
	t.Run("PartialMapTest", SetValueTest(values, "a.b[0].c[0]", true, false))
	t.Run("PartialMapTestNonFirstIntermediateNodes1", SetValueTest(values, "a.b[1].c[0]", true, true))
	t.Run("PartialMapTestNonFirstIntermediateNodes2", SetValueTest(values, "a.b[0].c[1]", true, false))

	values = map[string]interface{}{
		"a" : map[string]interface{}{
			"b":[]interface{}{
				map[string]interface{}{
					"c":[]interface{}{
						"c1",
					},
				},
			},
		},
	}
	t.Run("FullMapTestNoIntermediateNodes", SetValueTest(values, "a.b[0].c[0]", false, false))
	t.Run("FullMapTest", SetValueTest(values, "a.b[0].c[0]", true, false))
	t.Run("FullMapTestOutOfIndex1", SetValueTest(values, "a.b[1].c[0]", true, true))
	t.Run("FullMapTestOutOfIndex2", SetValueTest(values, "a.b[0].c[1]", true, true))
}

/*********************
	SubTests
 *********************/
func SetValueTest(values map[string]interface{}, key string, createIntermediateNodes bool, expectedFail bool) func(*testing.T) {
	return func(t *testing.T) {
		g := NewWithT(t)
		const testVal = "test value"
		values, _ = deepCopy(values)
		e := setValue(values, key, testVal, createIntermediateNodes)
		if expectedFail {
			g.Expect(e).To(HaveOccurred(), `setValue at "%s" should return error`, key)
		} else {
			g.Expect(e).To(Succeed(), `setValue at "%s" shouldn't return error`, key)
			g.Expect(value(values, key)).To(Equal(testVal), `value at "%s" should be correct`, key)
		}
	}
}

/*********************
	Helper
 *********************/
// Map performs a deep copy of the given map m.
func deepCopy(m map[string]interface{}) (map[string]interface{}, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	dec := gob.NewDecoder(&buf)
	if e := enc.Encode(m); e != nil {
		panic(e)
	}

	var cp map[string]interface{}
	if e := dec.Decode(&cp); e != nil {
		panic(e)
	}
	return cp, nil
}
