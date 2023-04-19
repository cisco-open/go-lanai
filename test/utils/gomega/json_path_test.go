package gomegautils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

const TestJson = `
{
    "firstName": "John",
    "lastName": "doe",
    "age": 26,
    "address": {
        "streetAddress": "naist street",
        "city": "Nara",
        "postalCode": "630-0192"
    },
    "phoneNumbers": [
        {
            "type": "iPhone",
            "number": "0123-4567-8888"
        },
        {
            "type": "home",
            "number": "0123-4567-8910"
        }
    ]
}
`

func TestJsonPathMatchers(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestHaveJsonPath(), "TestHaveJsonPath"),
		test.GomegaSubTest(SubTestHaveJsonPathWithValue(), "TestHaveJsonPathWithValue"),
		test.GomegaSubTest(SubTestFailureMessages(), "TestFailureMessages"),
	)
}

func SubTestHaveJsonPath() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(TestJson).To(HaveJsonPath("$.firstName"))
		g.Expect(TestJson).To(HaveJsonPath("$..type"))
		g.Expect(TestJson).NotTo(HaveJsonPath("$.type"))
	}
}

func SubTestHaveJsonPathWithValue() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(TestJson).To(HaveJsonPathWithValue("$.firstName", ContainElements("John")))
		g.Expect(TestJson).To(HaveJsonPathWithValue("$.lastName", "doe"))
		g.Expect(TestJson).To(HaveJsonPathWithValue("$..type", HaveLen(2)))
		g.Expect(TestJson).NotTo(HaveJsonPathWithValue("$.type", ContainElement("Android")))
	}
}

func SubTestFailureMessages() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		matcher := HaveJsonPathWithValue("$.lastName", "doe")
		msg := matcher.FailureMessage([]byte(TestJson))
		g.Expect(msg).To(Not(BeEmpty()))
		msg = matcher.NegatedFailureMessage(TestJson)
		g.Expect(msg).To(Not(BeEmpty()))
	}
}
