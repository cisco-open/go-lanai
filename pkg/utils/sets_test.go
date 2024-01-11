package utils

import (
	"context"
	"cto-github.cisco.com/NFV-BU/go-lanai/test"
	"encoding/json"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"testing"
	"time"
)

func TestSets(t *testing.T) {
	test.RunTest(context.Background(), t,
		test.GomegaSubTest(SubTestStringSet(), "TestStringSet"),
		test.GomegaSubTest(SubTestSet(), "TestSet"),
		test.GomegaSubTest(SubTestGenericSet(), "TestGenericSet"),
	)
}

/*************************
	Sub-Test Cases
 *************************/

func SubTestStringSet() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		set := NewStringSet("v1", "v2")
		AssertStringSet(g, set, "v1", "v2")

		cpy := set.Copy()
		AssertStringSet(g, cpy, "v1", "v2")
		g.Expect(cpy.Equals(set)).To(BeTrue(), "copy of set should equals to the original")

		fromStringSet := NewStringSetFrom(set)
		AssertStringSet(g, fromStringSet, "v1", "v2")

		fromSet := NewStringSetFrom(NewSet("v1", "v2", 2))
		AssertStringSet(g, fromSet, "v1", "v2")

		fromStringSlice := NewStringSetFrom([]string{"v1", "v2"})
		AssertStringSet(g, fromStringSlice, "v1", "v2")

		fromSlice := NewStringSetFrom([]interface{}{"v1", "v2"})
		AssertStringSet(g, fromSlice, "v1", "v2")
	}
}

func SubTestSet() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		now := time.Now()
		hashable := 5
		set := NewSet(now, hashable)
		AssertSet(g, set, now, hashable)

		cpy := set.Copy()
		AssertSet(g, cpy, now, hashable)
		g.Expect(cpy.Equals(set)).To(BeTrue(), "copy of set should equals to the original")

		fromStringSet := NewSetFrom(NewStringSet("v1", "v2"))
		AssertSet(g, fromStringSet, "v1", "v2")

		fromSet := NewSetFrom(set)
		AssertSet(g, fromSet, now, hashable)

		fromStringSlice := NewSetFrom([]string{"v1", "v2"})
		AssertSet(g, fromStringSlice, "v1", "v2")

		fromSlice := NewSetFrom([]interface{}{now, hashable})
		AssertSet(g, fromSlice, now, hashable)
	}
}

func SubTestGenericSet() test.GomegaSubTestFunc {
	return func(ctx context.Context, t *testing.T, g *gomega.WithT) {
		type Type struct {
			Value string
		}
		newFn := func() Type {
			return Type{Value: RandomString(5)}
		}
		v1, v2 := newFn(), newFn()
		set := NewGenericSet(v1, v2)
		AssertGenericSet(g, set, newFn, v1, v2)

		cpy := set.Copy()
		AssertGenericSet(g, cpy, newFn, v1, v2)
		g.Expect(cpy.Equals(set)).To(BeTrue(), "copy of set should equals to the original")
	}
}

/*************************
	Helpers
 *************************/

func AssertStringSet(g *gomega.WithT, set StringSet, expected...string) {
	g.Expect(set.HasAll(expected...)).To(BeTrue(), "HasAll should be correct")
	g.Expect(set.Values()).To(ContainElements(expected), "Values should be correct")
	for _, v := range expected {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correct", v)
	}
	g.Expect(set.Has("non-exists")).To(BeFalse(), "Has(%v) should be correct", "non-exists")

	extraValues := []string{"added1", "added2"}
	set.Add(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeTrue(), "HasAll should be correct after Add")
	g.Expect(set.Values()).To(ContainElements(extraValues), "Values should be correct after Add")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correctafter Add", v)
	}

	set.Remove(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeFalse(), "HasAll should be correct after Remove")
	g.Expect(set.Values()).ToNot(ContainElements(extraValues), "Values should be correct after Remove")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeFalse(), "Has(%v) should be correct after Remove", v)
	}

	data, e := json.Marshal(set)
	g.Expect(e).To(Succeed(), "JSON marshalling should not fail")
	var another StringSet
	e = json.Unmarshal(data, &another)
	g.Expect(e).To(Succeed(), "JSON unmarshalling should not fail")
	g.Expect(set.Equals(another)).To(BeTrue(), "unmarshalled set should equals to the original")
}

func AssertSet(g *gomega.WithT, set Set, expected...interface{}) {
	g.Expect(set.HasAll(expected...)).To(BeTrue(), "HasAll should be correct")
	g.Expect(set.Values()).To(ContainElements(expected), "Values should be correct")
	for _, v := range expected {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correct", v)
	}
	g.Expect(set.Has("non-exists")).To(BeFalse(), "Has(%s) should be correct", "non-exists")

	extraValues := []interface{}{true, 2.4}
	set.Add(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeTrue(), "HasAll should be correct after Add")
	g.Expect(set.Values()).To(ContainElements(extraValues), "Values should be correct after Add")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correctafter Add", v)
	}

	set.Remove(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeFalse(), "HasAll should be correct after Remove")
	g.Expect(set.Values()).ToNot(ContainElements(extraValues), "Values should be correct after Remove")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeFalse(), "Has(%v) should be correct after Remove", v)
	}

	data, e := json.Marshal(set)
	g.Expect(e).To(Succeed(), "JSON marshalling should not fail")
	var another Set
	e = json.Unmarshal(data, &another)
	g.Expect(e).To(Succeed(), "JSON unmarshalling should not fail")
}

func AssertGenericSet[T comparable](g *gomega.WithT, set GenericSet[T], newFn func() T, expected...T) {
	g.Expect(set.HasAll(expected...)).To(BeTrue(), "HasAll should be correct")
	g.Expect(set.Values()).To(ContainElements(expected), "Values should be correct")
	for _, v := range expected {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correct", v)
	}
	g.Expect(set.Has(newFn())).To(BeFalse(), "Has(%v) should be correct", "non-exists")

	extraValues := []T{newFn(), newFn()}
	set.Add(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeTrue(), "HasAll should be correct after Add")
	g.Expect(set.Values()).To(ContainElements(extraValues), "Values should be correct after Add")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeTrue(), "Has(%v) should be correctafter Add", v)
	}

	set.Remove(extraValues...)
	g.Expect(set.HasAll(extraValues...)).To(BeFalse(), "HasAll should be correct after Remove")
	g.Expect(set.Values()).ToNot(ContainElements(extraValues), "Values should be correct after Remove")
	for _, v := range extraValues {
		g.Expect(set.Has(v)).To(BeFalse(), "Has(%v) should be correct after Remove", v)
	}

	data, e := json.Marshal(set)
	g.Expect(e).To(Succeed(), "JSON marshalling should not fail")
	var another GenericSet[T]
	e = json.Unmarshal(data, &another)
	g.Expect(e).To(Succeed(), "JSON unmarshalling should not fail")
	g.Expect(set.Equals(another)).To(BeTrue(), "unmarshalled set should equals to the original")
}