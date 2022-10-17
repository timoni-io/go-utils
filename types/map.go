package types

import (
	"golang.org/x/exp/constraints"
)

type SortFunction[K constraints.Ordered] func(data []K, i, j int) bool

type Weighted[V any] struct {
	Value  V
	Weight uint32
}
