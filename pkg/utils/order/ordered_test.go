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

package order

import (
	"fmt"
	"github.com/google/uuid"
	. "github.com/onsi/gomega"
	"math/rand"
	"sort"
	"testing"
	"time"
)

func TestOrderedFirstCompare(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	expectFunc := func(p,o,r int) expectation {
		return expectation{
			priority: []int{0, p},
			ordered: []int{p, o},
			regular: []int{p+o, r},
			desc: false,
		}
	}
	t.Run("SortTest",
		SortTest(OrderedFirstCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("SortStableTest",
		SortStableTest(OrderedFirstCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("CompareFuncDirectUsageTest",
		CompareFuncDirectUsageTest(OrderedFirstCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
}

func TestOrderedFirstCompareReverse(t *testing.T) {
	expectFunc := func(p,o,r int) expectation {
		return expectation{
			priority: []int{r+o, p},
			ordered: []int{r, o},
			regular: []int{0, r},
			desc: true,
		}
	}
	t.Run("SortTest",
		SortTest(OrderedFirstCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("SortStableTest",
		SortStableTest(OrderedFirstCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("CompareFuncDirectUsageTest",
		CompareFuncDirectUsageTest(OrderedFirstCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
}

func TestOrderedLastCompare(t *testing.T) {
	expectFunc := func(p,o,r int) expectation {
		return expectation{
			priority: []int{r, p},
			ordered: []int{r+p, o},
			regular: []int{0, r},
			desc: false,
		}
	}
	t.Run("SortTest",
		SortTest(OrderedLastCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("SortStableTest",
		SortStableTest(OrderedLastCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("CompareFuncDirectUsageTest",
		CompareFuncDirectUsageTest(OrderedLastCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
}

func TestOrderedLastCompareReverse(t *testing.T) {
	expectFunc := func(p,o,r int) expectation {
		return expectation{
			priority: []int{o, p},
			ordered: []int{0, o},
			regular: []int{p+o, r},
			desc: true,
		}
	}
	t.Run("SortTest",
		SortTest(OrderedLastCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("SortStableTest",
		SortStableTest(OrderedLastCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("CompareFuncDirectUsageTest",
		CompareFuncDirectUsageTest(OrderedLastCompareReverse, randdomSize(), randdomSize(), randdomSize(), expectFunc))
}

func TestUnOrderedMiddleCompare(t *testing.T) {
	expectFunc := func(p,o,r int) expectation {
		return expectation{
			priority: []int{0, p},
			ordered: []int{p+r, o},
			regular: []int{p, r},
			desc: false,
		}
	}
	t.Run("SortTest",
		SortTest(UnorderedMiddleCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("SortStableTest",
		SortStableTest(UnorderedMiddleCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
	t.Run("CompareFuncDirectUsageTest",
		CompareFuncDirectUsageTest(UnorderedMiddleCompare, randdomSize(), randdomSize(), randdomSize(), expectFunc))
}


/**************************
	SubTests
 **************************/
type expectFunc func(p, o, r int) expectation

func SortTest(compareFunc CompareFunc, p, o, r int, expect expectFunc) func(t *testing.T) {
	return func(t *testing.T) {
		slice := makeRandomSlice(p, o, r)
		Sort(slice, compareFunc)
		assertSorted(t, slice, expect(p, o, r))
	}
}

func SortStableTest(compareFunc CompareFunc, p, o, r int, expect expectFunc) func(t *testing.T) {
	return func(t *testing.T) {
		slice := makeRandomSlice(p, o, r)
		SortStable(slice, compareFunc)
		assertStableSorted(t, slice, expect(p, o, r))
	}
}

func CompareFuncDirectUsageTest(compareFunc CompareFunc, p, o, r int, expect expectFunc) func(t *testing.T) {
	return func(t *testing.T) {
		slice := makeRandomSlice(p, o, r)
		sort.Slice(slice, func(i,j int) bool {
			return compareFunc(slice[i], slice[j])
		})
		assertSorted(t, slice, expect(p, o, r))
	}
}

/**************************
	Helpers
 **************************/
func assertStableSorted(t *testing.T, slice []unique, expected expectation) {
	g := NewWithT(t)
	// priority
	assertSortedPortion(g, slice, expected.priority[0], expected.priority[1], expected.desc,
		(*priorityItem)(nil), func(i interface{}) int {
			return i.(PriorityOrdered).PriorityOrder()
		}, func(i interface{}) int {
			return i.(unique).NaturalOrder()
		})

	// ordered
	assertSortedPortion(g, slice, expected.ordered[0], expected.ordered[1], expected.desc,
		(*orderedItem)(nil), func(i interface{}) int {
			return i.(Ordered).Order()
		}, func(i interface{}) int {
			return i.(unique).NaturalOrder()
		})

	// regular
	assertSortedPortion(g, slice, expected.regular[0], expected.regular[1], expected.desc,
		(*regularItem)(nil), func(i interface{}) int {
			return 0
		}, func(i interface{}) int {
			return i.(unique).NaturalOrder()
		})
}

func assertSorted(t *testing.T, slice []unique, expected expectation) {
	g := NewWithT(t)
	// priority
	assertSortedPortion(g, slice, expected.priority[0], expected.priority[1], expected.desc,
		(*priorityItem)(nil), func(i interface{}) int {
			return i.(PriorityOrdered).PriorityOrder()
		}, nil)

	// ordered
	assertSortedPortion(g, slice, expected.ordered[0], expected.ordered[1], expected.desc,
		(*orderedItem)(nil), func(i interface{}) int {
			return i.(Ordered).Order()
		}, nil)

	// regular
	assertSortedPortion(g, slice, expected.regular[0], expected.regular[1], expected.desc,
		(*regularItem)(nil), func(i interface{}) int {
			return 0
		}, nil)
}

func assertSortedPortion(g *WithT,
	slice []unique,
	start, size int, desc bool,
	expectedType interface{},
	orderExtractor func(interface{}) int,
	naturalOrderExtractor func(interface{}) int ) {

	for i := 0; i < size; i++ {
		idx := i + start
		g.Expect(idx < len(slice)).To(BeTrue(), "length should be greater than %d", idx)
		g.Expect(slice[idx]).
			To(BeAssignableToTypeOf(expectedType), "item at %d should be %T", idx, expectedType)
		if i == 0 {
			continue
		}
		// check order value if not first item
		if desc {
			g.Expect(orderExtractor(slice[idx]) <= orderExtractor(slice[idx-1])).
				To(BeTrue(), "item at %d should be %T and have order less than previous", idx, expectedType)
		} else {
			g.Expect(orderExtractor(slice[idx]) >= orderExtractor(slice[idx-1])).
				To(BeTrue(), "item at %d should be %T and have order greater than previous", idx, expectedType)
		}

		if naturalOrderExtractor != nil && orderExtractor(slice[idx]) == orderExtractor(slice[idx-1]) {
			g.Expect(naturalOrderExtractor(slice[idx]) >= naturalOrderExtractor(slice[idx-1])).
				To(BeTrue(), "item at %d (natural order: %d) should be %T and have natural order greater than previous (natural order: %d)", idx, naturalOrderExtractor(slice[idx]), expectedType, naturalOrderExtractor(slice[idx-1]))
		}
	}
}

func randdomSize() int {
	return rand.Int() % 50
}

func makeRandomSlice(priority, order, regular int) []unique {
	s := make([]unique, priority + order + regular)
	i := 0
	for j := 0; j < priority; j++ {
		s[i] = newPriority(rand.Int())
		i++
	}

	for j := 0; j < order; j++ {
		s[i] = newOrdered(rand.Int())
		i++
	}

	for j := 0; j < regular; j++ {
		s[i] = newRegular()
		i++
	}

	rand.Shuffle(len(s), func(i,j int) {
		s[i], s[j] = s[j], s[i]
	})

	for j, _ := range s {
		s[j].SetNaturalOrder(j)
	}

	return s
}

// used for assertion. contains index ranges of each type of item within a slice
// for each  []int contains two values, first is the starting index, second is length
type expectation struct {
	priority []int
	ordered  []int
	regular  []int
	desc     bool
}

type unique interface {
	Id() string
	NaturalOrder() int
	SetNaturalOrder(naturalOrder int)
}

// priorityItem implemnts PriorityOrdered, unique and Stringer
type priorityItem struct {
	order        int
	id           string
	naturalOrder int
}

func newPriority(order int) *priorityItem {
	return &priorityItem{
		order:        order,
		id:           uuid.New().String(),
	}
}

func (i priorityItem) Id() string {
	return i.id
}

func (i priorityItem) PriorityOrder() int {
	return i.order
}

func (i priorityItem) NaturalOrder() int {
	return i.naturalOrder
}

func (i *priorityItem) SetNaturalOrder(naturalOrder int) {
	i.naturalOrder = naturalOrder
}

func (i priorityItem) String() string {
	return fmt.Sprintf("p[%d]: %s", i.order, i.id)
}

// orderedItem implemnts Ordered, unique and Stringer
type orderedItem struct {
	order        int
	id           string
	naturalOrder int
}

func newOrdered(order int) *orderedItem {
	return &orderedItem{
		order:        order,
		id:           uuid.New().String(),
	}
}

func (i orderedItem) Id() string {
	return i.id
}

func (i orderedItem) Order() int {
	return i.order
}

func (i orderedItem) NaturalOrder() int {
	return i.naturalOrder
}

func (i *orderedItem) SetNaturalOrder(naturalOrder int) {
	i.naturalOrder = naturalOrder
}

func (i orderedItem) String() string {
	return fmt.Sprintf("[%d]: %s", i.order, i.id)
}

// regularItem only implemnts unique and Stringer
type regularItem struct {
	id           string
	naturalOrder int
}

func newRegular() *regularItem {
	return &regularItem{
		id:           uuid.New().String(),
	}
}

func (i regularItem) Id() string {
	return i.id
}

func (i regularItem) NaturalOrder() int {
	return i.naturalOrder
}

func (i *regularItem) SetNaturalOrder(naturalOrder int) {
	i.naturalOrder = naturalOrder
}

func (i regularItem) String() string {
	return fmt.Sprintf("----: %s", i.id)
}
