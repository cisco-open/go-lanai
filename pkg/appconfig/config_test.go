// Copyright 2023 Cisco Systems, Inc. and its affiliates
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package appconfig

import (
	"bytes"
	"encoding/gob"
	"fmt"
	. "github.com/onsi/gomega"
	"testing"
)

func init() {
	gob.Register(map[string]interface{}{})
	gob.Register([]interface{}{})
}

func TestParsePlaceHolders(t *testing.T) {
	// Without Default Values
	v := "${my.place.holder}"
	expected := map[string]interface{} {
		"my.place.holder": nil,
	}
	t.Run("PlaceholderPattern1", PlaceholderParsingTest(v, expected, false))

	v = "1${my.place.holder}"
	expected = map[string]interface{} {
		"my.place.holder": nil,
	}
	t.Run("PlaceholderPattern2", PlaceholderParsingTest(v, expected, true))

	v = "${my.place.holder}1"
	expected = map[string]interface{} {
		"my.place.holder": nil,
	}
	t.Run("PlaceholderPattern3", PlaceholderParsingTest(v, expected, true))

	v = "${my.place.holder}${my.second.holder}"
	expected = map[string]interface{} {
		"my.place.holder": nil,
		"my.second.holder": nil,
	}
	t.Run("PlaceholderPattern4", PlaceholderParsingTest(v, expected, true))

	// With Default Values
	v = "${my.place.holder:200}"
	expected = map[string]interface{} {
		"my.place.holder": 200.0,
	}
	t.Run("PlaceholderPattern5", PlaceholderParsingTest(v, expected, false))

	v = "1${my.place.holder:true}"
	expected = map[string]interface{} {
		"my.place.holder": true,
	}
	t.Run("PlaceholderPattern6", PlaceholderParsingTest(v, expected, true))

	v = "${my.place.holder:default_value}1"
	expected = map[string]interface{} {
		"my.place.holder": "default_value",
	}
	t.Run("PlaceholderPattern7", PlaceholderParsingTest(v, expected, true))

	v = "${my.place.holder:mixed}${my.second.holder:value}"
	expected = map[string]interface{} {
		"my.place.holder": "mixed",
		"my.second.holder": "value",
	}
	t.Run("PlaceholderPattern8", PlaceholderParsingTest(v, expected, true))

	// Invalid patterns
	v = "${${my.second.holder}}"
	expected = map[string]interface{} {
		"my.place.holder": nil,
	}
	t.Run("PlaceholderInvalidPattern1", PlaceholderFailedParsingTest(v))
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

func TestConfigBind(t *testing.T) {
	type fields struct {
		properties properties
		isLoaded   bool
	}
	type args struct {
		target interface{}
		prefix string
	}

	type testProperties struct {
		TestInnerProp bool
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		expectedErr error
		expected    map[string]interface{}
	}{
		{
			name: "Bind Should Err If Config Not Loaded",
			fields: fields{
				properties: map[string]interface{}{},
				isLoaded:   false,
			},
			args: args{
				target: nil,
				prefix: "",
			},
			expectedErr: errBindWithConfigBeforeLoaded,
			expected:    nil,
		},
		{
			name: "Bind Should Map Properties To Target",
			fields: fields{
				properties: map[string]interface{}{
					"test": testProperties{
						TestInnerProp: true,
					},
				},
				isLoaded: true,
			},
			args: args{
				target: testProperties{},
				prefix: "test",
			},
			expectedErr: nil,
			expected:    map[string]interface{}{"TestInnerProp": true},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &config{
				properties: tt.fields.properties,
				isLoaded:   tt.fields.isLoaded,
			}

			g := NewWithT(t)
			err := c.Bind(&tt.args.target, tt.args.prefix)
			if tt.expectedErr != nil {
				g.Expect(err).To(MatchError(tt.expectedErr), "Bind() error = %v, expectedErr = %v", err, tt.expectedErr)
			} else {
				g.Expect(err).To(Succeed(), `Bind() with %v should not have returned error`, c)
				g.Expect(tt.args.target).To(Equal(tt.expected), "Bind() target = %v, expectedResult = %v", tt.args.target, tt.expected)
			}
		})
	}
}

/*********************
	SubTests
 *********************/
func PlaceholderParsingTest(v string, expectedPlaceholders map[string]interface{}, expectEmbedded bool) func(*testing.T) {
	return func(t *testing.T) {
		g := NewWithT(t)
		placeholders, isEmbedded, e := parsePlaceHolder(v)
		g.Expect(e).To(Succeed(), "parsing should not returns error")
		g.Expect(placeholders).To(HaveLen(len(expectedPlaceholders)), "number of placeholders should be correct")
		g.Expect(isEmbedded).To(Equal(expectEmbedded), "isEmbedded should be correct")
		for _, ph := range placeholders {
			g.Expect(expectedPlaceholders).To(HaveKey(ph.key), "placeholder [%s] should have correct key", ph.key)
			dv, _ := expectedPlaceholders[ph.key]
			if dv != nil {
				g.Expect(ph.defaultVal).To(Equal(dv), "placeholder [%s] should have correct default value [%v]", ph.key, dv)
			} else {
				g.Expect(ph.defaultVal).To(BeNil(), "placeholder [%s] should have nil default value [%v]", ph.key, dv)
			}
		}
	}
}

func PlaceholderFailedParsingTest(v string) func(*testing.T) {
	return func(t *testing.T) {
		g := NewWithT(t)
		_, _, e := parsePlaceHolder(v)
		g.Expect(e).To(Not(Succeed()), "parsing should returns error")
	}
}

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
