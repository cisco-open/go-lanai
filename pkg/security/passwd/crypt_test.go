package passwd_test

import (
    "context"
    "github.com/cisco-open/go-lanai/pkg/security/passwd"
    "github.com/cisco-open/go-lanai/test"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "testing"
)

/*************************
	Test
 *************************/

func TestPasswordEncoder(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestBCryptPasswordEncoder(), "BCryptPasswordEncoder"),
		test.GomegaSubTest(SubTestNoopPasswordEncoder(), "NoopPasswordEncoder"),
	)
}

/*************************
	Sub Tests
 *************************/

func SubTestBCryptPasswordEncoder() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        const password = `test-password`
        const bcrypt12 = `$2y$12$A4UQV/bbA1XOKADxI/SRCeyGsBsxiT42efCeXzTd/.LbKVVgJfB92`
        const bcrypt10 = `$2y$12$/UE0PRYdLLzqsNU3Y6aNWeDUyveQAvWAdfJPJ9PdNO3Oh23yPlXKC`
        const wrongBcrypt = `$2y$08$UxIFO8VcBVXN8g2SOm9wl.ZCFs4X5qVTdeJZsYq.7OhXGvxP8ZWfG`

        encoder := passwd.NewBcryptPasswordEncoder()
        encoded := encoder.Encode(password)
        g.Expect(encoded).ToNot(BeEmpty(), "encoded password should not be empty")
        g.Expect(encoded).ToNot(Equal(password), "encoded password should not be same as raw password")

        ok := encoder.Matches(password, encoded)
        g.Expect(ok).To(BeTrue(), "encoded password should match raw password")
        ok = encoder.Matches(password, bcrypt10)
        g.Expect(ok).To(BeTrue(), "encoded password from other tools should match raw password")
        ok = encoder.Matches(password, bcrypt12)
        g.Expect(ok).To(BeTrue(), "encoded password from other tools with different cost should match raw password")

        ok = encoder.Matches(password, wrongBcrypt)
        g.Expect(ok).To(BeFalse(), "wrong encoded password from other tools with different cost should not match raw password")
        ok = encoder.Matches(password, "malformed")
        g.Expect(ok).To(BeFalse(), "malformed password should not match raw password")
	}
}

func SubTestNoopPasswordEncoder() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        const password = `test-password`

        encoder := passwd.NewNoopPasswordEncoder()
        encoded := encoder.Encode(password)
        g.Expect(encoded).ToNot(BeEmpty(), "encoded password should not be empty")
        g.Expect(encoded).To(Equal(password), "encoded password should be same as raw password")

        ok := encoder.Matches(password, encoded)
        g.Expect(ok).To(BeTrue(), "encoded password should match raw password")

        ok = encoder.Matches(password, "wrong password")
        g.Expect(ok).To(BeFalse(), "wrong encoded password from other tools with different cost should not match raw password")
    }
}
