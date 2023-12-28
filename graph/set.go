package graph

import (
	"maps"
	"slices"
)

type SetAccess[T comparable] interface {
	Has(item T) bool
	Len() int
}

type insertOrderedSet[T comparable] struct {
	order []T
	set   map[T]struct{}
}

var _ SetAccess[int] = (*insertOrderedSet[int])(nil)

func newInsertOrderedSet[T comparable]() *insertOrderedSet[T] {
	return &insertOrderedSet[T]{
		set: map[T]struct{}{},
	}
}

func (set *insertOrderedSet[T]) CloneDeep() *insertOrderedSet[T] {
	return &insertOrderedSet[T]{
		order: slices.Clone(set.order),
		set:   maps.Clone(set.set),
	}
}

func (set *insertOrderedSet[T]) Len() int { return len(set.order) }

func (set *insertOrderedSet[T]) Has(item T) bool {
	_, ok := set.set[item]
	return ok
}

func (set *insertOrderedSet[T]) Add(item T) {
	if _, ok := set.set[item]; ok {
		return
	}
	set.order = append(set.order, item)
	set.set[item] = struct{}{}
}

func (set *insertOrderedSet[T]) Remove(item T) {
	if _, ok := set.set[item]; !ok {
		return
	}
	idx := slices.Index(set.order, item)
	if idx >= 0 {
		delete(set.set, item)
		set.order = slices.Delete(set.order, idx, idx+1)
	}
}
