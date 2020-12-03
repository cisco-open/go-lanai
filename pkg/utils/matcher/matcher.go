package matcher

import (
	"context"
)

type Matcher interface {
	Matches(interface{}) (bool, error)
	MatchesWithContext(context.Context, interface{}) (bool, error)
}

type ChainableMatcher interface {
	Matcher
	// ChainableMatcher.Or(...Matcher) concat given matchers with OR operator
	Or(matcher ...Matcher) ChainableMatcher
	// ChainableMatcher.And(...Matcher) concat given matchers with AND operator
	And(matcher ...Matcher) ChainableMatcher
}

// Any() returns a matcher that matches everything
func Any() ChainableMatcher {
	return NoopMatcher(true)
}

// None() returns a matcher that matches nothing
func None() ChainableMatcher {
	return NoopMatcher(false)
}

// Or(...) concat given matchers with OR operator
func Or(left Matcher, right...Matcher) ChainableMatcher {
	return OrMatcher(append([]Matcher{left}, right...))
}

// And(...) concat given matchers with AND operator
func And(left Matcher, right...Matcher) ChainableMatcher {
	return AndMatcher(append([]Matcher{left}, right...))
}

// Not(Matcher) returns a negated matcher
func Not(matcher Matcher) ChainableMatcher {
	return &NegateMatcher{matcher}
}

// NoopMatcher matches stuff literally
type NoopMatcher bool

func (m NoopMatcher) Matches(_ interface{}) (bool, error) {
	return bool(m), nil
}

func (m NoopMatcher) MatchesWithContext(context.Context, interface{}) (bool, error) {
	return bool(m), nil
}

func (m NoopMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m NoopMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

func (m NoopMatcher) String() string {
	if m {
		return "matches any"
	} else {
		return "matches none"
	}
}

// OrMatcher chain a list of matchers with OR operator
type OrMatcher []Matcher

func (m OrMatcher) Matches(i interface{}) (ret bool, err error) {
	for _,item := range m {
		if ret,err = item.Matches(i); ret || err != nil {
			break
		}
	}
	return
}

func (m OrMatcher) MatchesWithContext(c context.Context, i interface{}) (ret bool, err error) {
	for _,item := range m {
		if ret,err = item.MatchesWithContext(c, i); ret || err != nil {
			break
		}
	}
	return
}

func (m OrMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m OrMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

// AndMatcher chain a list of matchers with AND operator
type AndMatcher []Matcher

func (m AndMatcher) Matches(i interface{}) (ret bool, err error) {
	for _,item := range m {
		if ret,err = item.Matches(i); !ret || err != nil {
			break
		}
	}
	return
}

func (m AndMatcher) MatchesWithContext(c context.Context, i interface{}) (ret bool, err error) {
	for _,item := range m {
		if ret,err = item.MatchesWithContext(c, i); !ret || err != nil {
			break
		}
	}
	return
}

func (m AndMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m AndMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

// NegateMatcher apply ! operator to embedded Matcher
type NegateMatcher struct {
	Matcher
}

func (m *NegateMatcher) Matches(i interface{}) (ret bool, err error) {
	ret, err = m.Matcher.Matches(i)
	return !ret, err
}

func (m NegateMatcher) MatchesWithContext(c context.Context, i interface{}) (ret bool, err error) {
	ret, err = m.Matcher.MatchesWithContext(c, i)
	return !ret, err
}

func (m *NegateMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m *NegateMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}

// TODO review use cases to determine if this class is necessary
// GenericMatcher implements ChainableMatcher
type GenericMatcher struct {
	matchFunc func(context.Context, interface{}) (bool, error)
}

func (m *GenericMatcher) Matches(i interface{}) (bool, error) {
	return m.matchFunc(context.TODO(), i)
}

func (m GenericMatcher) MatchesWithContext(c context.Context, i interface{}) (ret bool, err error) {
	return m.matchFunc(c, i)
}

func (m *GenericMatcher) Or(matchers ...Matcher) ChainableMatcher {
	return Or(m, matchers...)
}

func (m *GenericMatcher) And(matchers ...Matcher) ChainableMatcher {
	return And(m, matchers...)
}