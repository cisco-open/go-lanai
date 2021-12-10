package order

import (
	"fmt"
	"reflect"
	"sort"
)

const (
	Lowest  = int(^uint(0) >> 1) // max int
	Highest = -Lowest - 1        // min int
)

type Ordered interface {
	Order() int
}

type PriorityOrdered interface {
	PriorityOrder() int
}

// LessFunc is accepted less func by sort.Slice and sort.SliceStable
type LessFunc func(i, j int) bool

// CompareFunc is used to compare two interface's order,
type CompareFunc func(left interface{}, right interface{}) bool

// Sort wraps sort.Slice with LessFunc constructed from given CompareFunc using reflect
// function panic if given interface is not slice
func Sort(slice interface{}, compareFunc CompareFunc) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Errorf("Sort only support slice, but got %T", slice))
	}
	sort.Slice(slice, func(i, j int) bool {
		return compareFunc(sv.Index(i).Interface(), sv.Index(j).Interface())
	})
}

// SortStable wraps sort.SliceStable with LessFunc constructed from given CompareFunc using reflect
// function panic if given interface is not slice
func SortStable(slice interface{}, compareFunc CompareFunc) {
	sv := reflect.ValueOf(slice)
	if sv.Kind() != reflect.Slice {
		panic(fmt.Errorf("Sort only support slice, but got %T", slice))
	}
	sort.SliceStable(slice, func(i, j int) bool {
		return compareFunc(sv.Index(i).Interface(), sv.Index(j).Interface())
	})
}

// OrderedFirstCompare compares objects based on order interfaces with following rule
// - PriorityOrdered wins over other types
// - Ordered wins over non- PriorityOrdered
// - Same category will compare its corresponding order value
func OrderedFirstCompare(left interface{}, right interface{}) bool {
	// first consider PriorityOrder
	lp, lpok := left.(PriorityOrdered)
	rp, rpok := right.(PriorityOrdered)
	lo, look := left.(Ordered)
	ro, rook := right.(Ordered)

	switch {
	// PriorityOrdered cases
	case lpok && rpok:
		return lp.PriorityOrder() < rp.PriorityOrder()
	case lpok && !rpok:
		return true
	case !lpok && rpok:
		return false
	// Ordered cases
	case look && rook:
		return lo.Order() < ro.Order()
	case look && !rook:
		return true
	case !look && rook:
		return false
	// not Ordered nor PriorityOrdered
	default:
		return false
	}
}

// OrderedFirstCompareReverse compares objects based on order interfaces with same rule as OrderedFirstCompare but reversed
func OrderedFirstCompareReverse(left interface{}, right interface{}) bool {
	return !OrderedFirstCompare(left, right)
}

// OrderedLastCompare compares objects based on order interfaces with following rule
// - Regular object (neither PriorityOrdered nor Ordered) wins over other types
// - PriorityOrdered wins over Ordered
// - Same category will compare its corresponding order value
func OrderedLastCompare(left interface{}, right interface{}) bool {
	// first consider PriorityOrder
	lp, lpok := left.(PriorityOrdered)
	rp, rpok := right.(PriorityOrdered)
	lo, look := left.(Ordered)
	ro, rook := right.(Ordered)

	switch {
	// left or right are regular object
	case !lpok && !look:
		return true
	case !rpok && !rook:
		return false
	// PriorityOrdered cases
	case lpok && rpok:
		return lp.PriorityOrder() < rp.PriorityOrder()
	case lpok && !rpok:
		return true
	case !lpok && rpok:
		return false
	// Ordered cases
	case look && rook:
		return lo.Order() < ro.Order()
	default:
		return false
	}
}

// OrderedLastCompareReverse compares objects based on order interfaces with same rule as OrderedLastCompare but reversed
func OrderedLastCompareReverse(left interface{}, right interface{}) bool {
	return !OrderedLastCompare(left, right)
}

// UnorderedMiddleCompare compares objects based on order interfaces with following rule
// - PriorityOrdered wins over other types
// - Regular object (neither PriorityOrdered nor Ordered) wins Ordered
// - Ordered at last
// - Same category will compare its corresponding order value
func UnorderedMiddleCompare(left interface{}, right interface{}) bool {
	// first consider PriorityOrder
	lp, lpok := left.(PriorityOrdered)
	rp, rpok := right.(PriorityOrdered)
	lo, look := left.(Ordered)
	ro, rook := right.(Ordered)

	switch {
	// PriorityOrdered cases
	case lpok && rpok:
		return lp.PriorityOrder() < rp.PriorityOrder()
	case lpok && !rpok:
		return true
	case !lpok && rpok:
		return false
	// Unordered case (not Ordered nor PriorityOrdered)
	case !look:
		return true
	case !rook:
		return false
	default:
		return lo.Order() < ro.Order()
	}
}

// UnorderedMiddleCompareReverse compares objects based on order interfaces with same rule as UnorderedMiddleCompare but reversed
func UnorderedMiddleCompareReverse(left interface{}, right interface{}) bool {
	return !UnorderedMiddleCompare(left, right)
}