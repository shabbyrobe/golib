package graph

type Vertex[T ID] struct {
	V T

	// Use an ordered set for these to keep execution order deterministic:
	out *insertOrderedSet[T] // Outward connections (dependencies) from this vertex to others
	in  *insertOrderedSet[T] // Inward connections (dependents) from other vertices into this one
}

func (v Vertex[T]) In() SetAccess[T]  { return v.in }
func (v Vertex[T]) Out() SetAccess[T] { return v.out }

func (v Vertex[T]) Indegree() int  { return v.in.Len() }
func (v Vertex[T]) Outdegree() int { return v.out.Len() }

func (v Vertex[T]) Empty() bool {
	return v.Indegree() == 0 && v.Outdegree() == 0
}

func (v Vertex[T]) CloneDeep() Vertex[T] {
	return Vertex[T]{
		V:   v.V,
		out: v.out.CloneDeep(),
		in:  v.in.CloneDeep(),
	}
}
