package cmdutils

import (
	"context"
	"fmt"
	"strings"
)

/*
   Experimental interfaces that intended to replace utils/matcher/* interfaces with generics
*/

type Matcher[T any] interface {
	Matches(T) (bool, error)
	MatchesWithContext(context.Context, T) (bool, error)
}

type ChainableMatcher[T any] interface {
	Matcher[T]
	// Or concat given matchers with OR operator
	Or(matcher ...Matcher[T]) ChainableMatcher[T]
	// And concat given matchers with AND operator
	And(matcher ...Matcher[T]) ChainableMatcher[T]
}

// Any returns a matcher that matches everything
func Any[T any]() ChainableMatcher[T] {
	return NoopMatcher[T](true)
}

// None returns a matcher that matches nothing
func None[T any]() ChainableMatcher[T] {
	return NoopMatcher[T](false)
}

// Or concat given matchers with OR operator
func Or[T any](left Matcher[T], right ...Matcher[T]) ChainableMatcher[T] {
	return OrMatcher[T](append([]Matcher[T]{left}, right...))
}

// And concat given matchers with AND operator
func And[T any](left Matcher[T], right ...Matcher[T]) ChainableMatcher[T] {
	return AndMatcher[T](append([]Matcher[T]{left}, right...))
}

// Not returns a negated matcher
func Not[T any](matcher Matcher[T]) ChainableMatcher[T] {
	return &NegateMatcher[T]{matcher}
}

// NoopMatcher matches stuff literally
type NoopMatcher[T any] bool

func (m NoopMatcher[T]) Matches(_ T) (bool, error) {
	return bool(m), nil
}

func (m NoopMatcher[T]) MatchesWithContext(context.Context, T) (bool, error) {
	return bool(m), nil
}

func (m NoopMatcher[T]) Or(matchers ...Matcher[T]) ChainableMatcher[T] {
	return Or[T](m, matchers...)
}

func (m NoopMatcher[T]) And(matchers ...Matcher[T]) ChainableMatcher[T] {
	return And[T](m, matchers...)
}

func (m NoopMatcher[T]) String() string {
	if m {
		return "matches any"
	} else {
		return "matches none"
	}
}

// OrMatcher chain a list of matchers with OR operator
type OrMatcher[T any] []Matcher[T]

func (m OrMatcher[T]) Matches(i T) (ret bool, err error) {
	for _, item := range m {
		if ret, err = item.Matches(i); ret || err != nil {
			break
		}
	}
	return
}

func (m OrMatcher[T]) MatchesWithContext(c context.Context, i T) (ret bool, err error) {
	for _, item := range m {
		if ret, err = item.MatchesWithContext(c, i); ret || err != nil {
			break
		}
	}
	return
}

func (m OrMatcher[T]) Or(matchers ...Matcher[T]) ChainableMatcher[T] {
	return Or[T](m, matchers...)
}

func (m OrMatcher[T]) And(matchers ...Matcher[T]) ChainableMatcher[T] {
	return And[T](m, matchers...)
}

func (m OrMatcher[T]) String() string {
	descs := make([]string, len(m))
	for i, item := range m {
		descs[i] = item.(fmt.Stringer).String()
	}
	return strings.Join(descs, " OR ")
}

// AndMatcher chain a list of matchers with AND operator
type AndMatcher[T any] []Matcher[T]

func (m AndMatcher[T]) Matches(i T) (ret bool, err error) {
	for _, item := range m {
		if ret, err = item.Matches(i); !ret || err != nil {
			break
		}
	}
	return
}

func (m AndMatcher[T]) MatchesWithContext(c context.Context, i T) (ret bool, err error) {
	for _, item := range m {
		if ret, err = item.MatchesWithContext(c, i); !ret || err != nil {
			break
		}
	}
	return
}

func (m AndMatcher[T]) Or(matchers ...Matcher[T]) ChainableMatcher[T] {
	return Or[T](m, matchers...)
}

func (m AndMatcher[T]) And(matchers ...Matcher[T]) ChainableMatcher[T] {
	return And[T](m, matchers...)
}

func (m AndMatcher[T]) String() string {
	descs := make([]string, len(m))
	for i, item := range m {
		descs[i] = item.(fmt.Stringer).String()
	}
	return strings.Join(descs, " AND ")
}

// NegateMatcher apply ! operator to embedded Matcher
type NegateMatcher[T any] struct {
	Matcher[T]
}

func (m NegateMatcher[T]) Matches(i T) (ret bool, err error) {
	ret, err = m.Matcher.Matches(i)
	return !ret, err
}

func (m NegateMatcher[T]) MatchesWithContext(c context.Context, i T) (ret bool, err error) {
	ret, err = m.Matcher.MatchesWithContext(c, i)
	return !ret, err
}

func (m NegateMatcher[T]) Or(matchers ...Matcher[T]) ChainableMatcher[T] {
	return Or[T](m, matchers...)
}

func (m NegateMatcher[T]) And(matchers ...Matcher[T]) ChainableMatcher[T] {
	return And[T](m, matchers...)
}

func (m NegateMatcher[T]) String() string {
	return fmt.Sprintf("Not(%v)", m.Matcher)
}

// GenericMatcher implements ChainableMatcher
// Implementing structs could directly use or embed this struct for convenience
type GenericMatcher[T any] struct {
	Description string
	MatchFunc   func(context.Context, T) (bool, error)
}

func (m *GenericMatcher[T]) Matches(i T) (bool, error) {
	return m.MatchFunc(context.TODO(), i)
}

func (m *GenericMatcher[T]) MatchesWithContext(c context.Context, i T) (ret bool, err error) {
	return m.MatchFunc(c, i)
}

func (m *GenericMatcher[T]) Or(matchers ...Matcher[T]) ChainableMatcher[T] {
	return Or(Matcher[T](m), matchers...)
}

func (m *GenericMatcher[T]) And(matchers ...Matcher[T]) ChainableMatcher[T] {
	return And(Matcher[T](m), matchers...)
}

func (m *GenericMatcher[T]) String() string {
	if len(m.Description) != 0 {
		return m.Description
	}
	return fmt.Sprintf("generic matcher with func [%T]", m.MatchFunc)
}
