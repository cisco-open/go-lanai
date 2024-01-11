package utils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestStringUtils(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestUnQuote(), "TestUnQuote"),
		test.GomegaSubTest(SubTestParseString(), "TestParseString"),
		test.GomegaSubTest(SubTestCamelToSnakeCase(), "TestCamelToSnakeCase"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestUnQuote() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		texts := map[string]string{
			` "double quoted"`:        `double quoted`,
			`'single quoted' `:        `single quoted`,
			`'John's single quoted'`: `John's single quoted`,
			` "`: `"`,
			`' `: `'`,
		}

		for text, expect := range texts {
			rs := UnQuote(text)
			g.Expect(rs).To(Equal(expect), "unquoted %s should be correct", text)
		}
	}
}

func SubTestParseString() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		texts := map[string]interface{}{
			`string`:                 "string",
			`123`:                    float64(123),
			`231.32`:                 231.32,
			`true`:                   true,
			`false`:                  false,
			`["string", 123, false]`: []interface{}{"string", float64(123), false},
			`{"string":"string", "num":123, "bool":true}`: map[string]interface{}{
				"string": "string",
				"num":    float64(123),
				"bool":   true,
			},
			`"not":"supported", "strings"`: `"not":"supported", "strings"`,
		}

		for text, expect := range texts {
			rs := ParseString(text)
			g.Expect(rs).To(BeAssignableToTypeOf(expect), "parsed string %s should be correct type", text)
			g.Expect(rs).To(BeEquivalentTo(expect), "parsed string %s should be correct value", text)
		}
	}
}

func SubTestCamelToSnakeCase() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		texts := map[string]string{
			`thisIsCamelCase`:    `this-is-camel-case`,
			`ThisIsNotCamelCase`: `this-is-not-camel-case`,
			`simple`:             `simple`,
			`Capital`:            `capital`,
			`CAPITAL`:            `capital`,
			`BaseURLValue`:       `base-urlvalue`,
			`baseURLvalue`:       `base-urlvalue`,
		}

		for text, expect := range texts {
			rs := CamelToSnakeCase(text)
			g.Expect(rs).To(Equal(expect), "CamelToSnakeCase %s should be correct", text)
		}
	}
}
