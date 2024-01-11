package utils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"errors"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
)

func TestRecoverableFunc(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestWithErrorReturn(), "TestWithErrorReturn"),
		test.GomegaSubTest(SubTestWithoutErrorReturn(), "TestWithoutErrorReturn"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestWithErrorReturn() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const errMsg = `oops`
		var e error
		e = RecoverableFunc(func() error {
			return errors.New(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function fails")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			panic(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			panic(errors.New(errMsg))
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() error {
			return nil
		})()
		g.Expect(e).To(Succeed(), "function should not fail when original function succeeded")
	}
}

func SubTestWithoutErrorReturn() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		const errMsg = `oops`
		var e error
		e = RecoverableFunc(func() {
			panic(errMsg)
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() {
			panic(errors.New(errMsg))
		})()
		g.Expect(e).To(HaveOccurred(), "function should fail when original function panic")
		g.Expect(e.Error()).To(Equal(errMsg), "error message should be correct")

		e = RecoverableFunc(func() {
			return
		})()
		g.Expect(e).To(Succeed(), "function should not fail when original function succeeded")
	}
}