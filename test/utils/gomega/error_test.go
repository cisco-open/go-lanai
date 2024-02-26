package gomegautils

import (
    "context"
    "errors"
    errorutils "github.com/cisco-open/go-lanai/pkg/utils/error"
    "github.com/cisco-open/go-lanai/test"
    "github.com/onsi/gomega"
    . "github.com/onsi/gomega"
    "testing"
)

const (
    TestReserved = 35 << errorutils.ReservedOffset
)

const (
    _                  = iota
    ErrorTypeCodeTestA = TestReserved + iota<<errorutils.ErrorTypeOffset
    ErrorTypeCodeTestB
)

const (
    _                        = iota
    ErrorSubTypeCodeA1 = ErrorTypeCodeTestA + iota<<errorutils.ErrorSubTypeOffset
    ErrorSubTypeCodeA2
)

const (
    _                                  = iota
    ErrorCodeA1A = ErrorSubTypeCodeA1 + iota
    ErrorCodeA1B
)

func TestIsErrorMatchers(t *testing.T) {
    test.RunTest(context.Background(), t,
        test.GomegaSubTest(SubTestIsError(), "SubTestIsError"),
        test.GomegaSubTest(SubTestHaveErrorType(), "SubTestHaveErrorType"),
        test.GomegaSubTest(SubTestHaveErrorSubType(), "SubTestHaveErrorSubType"),
        test.GomegaSubTest(SubTestHaveErrorCode(), "SubTestHaveErrorCode"),
        test.GomegaSubTest(SubTestErrorMatcherFailureMessages(), "SubTestErrorMatcherFailureMessages"),
    )
}

func SubTestIsError() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        nonCoded := errors.New("test error")
        concreteErr := errorutils.NewCodedError(ErrorCodeA1A, "concrete error")
        subTypeErr := errorutils.NewErrorSubType(ErrorSubTypeCodeA1, "subtype A1")
        typeErr := errorutils.NewErrorType(ErrorTypeCodeTestA, "type A")

        var e error
        e = wrap(nonCoded)
        g.Expect(e).To(IsError(nonCoded))
        g.Expect(e).ToNot(IsError(concreteErr))
        g.Expect(e).ToNot(IsError(subTypeErr))
        g.Expect(e).ToNot(IsError(typeErr))

        e = wrap(errorutils.NewCodedError(ErrorCodeA1A, "actual error"))
        g.Expect(e).ToNot(IsError(nonCoded))
        g.Expect(e).To(IsError(concreteErr))
        g.Expect(e).To(IsError(subTypeErr))
        g.Expect(e).To(IsError(typeErr))
    }
}

func SubTestHaveErrorType() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var e error
        e = errors.New("test error")
        g.Expect(e).ToNot(HaveErrorTypeCode(ErrorTypeCodeTestA))

        e = wrap(errorutils.NewCodedError(ErrorCodeA1A, "actual error"))
        g.Expect(e).To(HaveErrorTypeCode(ErrorTypeCodeTestA))
        g.Expect(e).ToNot(HaveErrorTypeCode(ErrorTypeCodeTestB))
    }
}

func SubTestHaveErrorSubType() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var e error
        e = errors.New("test error")
        g.Expect(e).ToNot(HaveErrorSubTypeCode(ErrorSubTypeCodeA1))

        e = wrap(errorutils.NewCodedError(ErrorCodeA1A, "actual error"))
        g.Expect(e).To(HaveErrorSubTypeCode(ErrorSubTypeCodeA1))
        g.Expect(e).ToNot(HaveErrorSubTypeCode(ErrorSubTypeCodeA2))
    }
}

func SubTestHaveErrorCode() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        var e error
        e = errors.New("test error")
        g.Expect(e).ToNot(HaveErrorCode(ErrorCodeA1A))

        e = wrap(errorutils.NewCodedError(ErrorCodeA1A, "actual error"))
        g.Expect(e).To(HaveErrorCode(ErrorCodeA1A))
        g.Expect(e).ToNot(HaveErrorCode(ErrorCodeA1B))
    }
}

func SubTestErrorMatcherFailureMessages() test.GomegaSubTestFunc {
    return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
        matcher := IsError(errorutils.NewErrorSubType(ErrorSubTypeCodeA1, "sub type A"))
        e := wrap(errorutils.NewCodedError(ErrorCodeA1A, "actual error"))
        msg := matcher.FailureMessage(e)
        g.Expect(msg).To(Not(BeEmpty()))
        msg = matcher.NegatedFailureMessage(TestJson)
        g.Expect(msg).To(Not(BeEmpty()))
    }
}

func wrap(e error) error {
    return wrapped{e}
}

type wrapped struct {
    error
}

func (w wrapped) Unwrap() error {
    return w.error
}