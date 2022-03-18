package test

import (
	"context"
	"github.com/onsi/gomega"
	"testing"
)

func TestWithSubTests(t *testing.T) {
	RunTest(context.Background(), t,
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-2"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-3"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-4"),
	)
}

func TestWithMixedResults(t *testing.T) {
	t.Skipf("Skipped because this test is meant to fail")
	RunTest(context.Background(), t,
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-1"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-1"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-2"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-2"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-3"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-3"),
		GomegaSubTest(SubTestAlwaysSucceed(), "SuccessfulTest-4"),
		GomegaSubTest(SubTestAlwaysFail(), "FailedTest-4"),
	)
}


func SubTestAlwaysSucceed() GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(true).To(gomega.BeTrue())
	}
}

func SubTestAlwaysFail() GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		g.Expect(true).To(gomega.BeFalse())
	}
}