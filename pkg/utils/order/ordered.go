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
	// first consider PriorityOrder
	lp, lpok := left.(PriorityOrdered)
	rp, rpok := right.(PriorityOrdered)
	lo, look := left.(Ordered)
	ro, rook := right.(Ordered)

	switch {
	// PriorityOrdered cases
	case lpok && rpok:
		return lp.PriorityOrder() > rp.PriorityOrder()
	case lpok && !rpok:
		return false
	case !lpok && rpok:
		return true
	// Ordered cases
	case look && rook:
		return lo.Order() > ro.Order()
	case look && !rook:
		return false
	case !look && rook:
		return true
	// not Ordered nor PriorityOrdered
	default:
		return false
	}
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
	// if both side are regular objects, there's no order
	case !lpok && !look && !rpok && !rook:
		return false
	// from here down, at least one side is not a regular object
	// left or right are regular object
	case !lpok && !look: //left is regular object
		return true
	case !rpok && !rook: //right is regular object
		return false
	// from here down, both side are ordered or priority ordered
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
		return false // theoretically wouldn't get here
	}
}

// OrderedLastCompareReverse compares objects based on order interfaces with same rule as OrderedLastCompare but reversed
func OrderedLastCompareReverse(left interface{}, right interface{}) bool {
	// first consider PriorityOrder
	lp, lpok := left.(PriorityOrdered)
	rp, rpok := right.(PriorityOrdered)
	lo, look := left.(Ordered)
	ro, rook := right.(Ordered)

	switch {
	// if both side are regular objects, there's no order
	case !lpok && !look && !rpok && !rook:
		return false
	// from here down, at least one side is not a regular object
	// left or right are regular object
	case !lpok && !look: //left is regular object
		return false
	case !rpok && !rook: //right is regular object
		return true
	// from here down, both side are ordered or priority ordered
	// PriorityOrdered cases
	case lpok && rpok:
		return lp.PriorityOrder() > rp.PriorityOrder()
	case lpok && !rpok:
		return false
	case !lpok && rpok:
		return true
	// Ordered cases
	case look && rook:
		return lo.Order() > ro.Order()
	default:
		return false //theoretically wouldn't get here
	}
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
	// PriorityOrdered cases - if at least one of the operand is PriorityOrdered
	case lpok && rpok:
		return lp.PriorityOrder() < rp.PriorityOrder()
	case lpok && !rpok:
		return true
	case !lpok && rpok:
		return false
	// Both are unordered - at this point, we know they are not PriorityOrdered, so we just need to check both are also not Ordered
	case !look && !rook:
		return false // return false to indicate left is not less than right, so that the natural order is kept
	// Left operand is not ordered, right operand is ordered
	// return true so that un ordered comes before ordered
	case !look:
		return true
	// Right operand is not ordered, left operand is ordered
	// return false so that un ordered comes before ordered
	case !rook:
		return false
	// both side are ordered
	default:
		return lo.Order() < ro.Order()
	}
}

// UnorderedMiddleCompareReverse compares objects based on order interfaces with same rule as UnorderedMiddleCompare but reversed
func UnorderedMiddleCompareReverse(left interface{}, right interface{}) bool {
	return !UnorderedMiddleCompare(left, right)
}
