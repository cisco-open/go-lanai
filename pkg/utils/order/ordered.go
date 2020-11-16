package order

const (
	Lowest = int(^uint(0) >> 1) // max int
	Highest = -Lowest - 1 // min int
)

type Ordered interface {
	Order() int
}

type PriorityOrdered interface {
	PriorityOrder() int
}

// ComparatorFunc is used for sort to compare two interface's order
type ComparatorFunc func(left interface{}, right interface{}) bool

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