package graph

import (
	"maps"
	"slices"
	"sort"
)

type ID interface {
	comparable
}

type Edge[T ID] struct {
	From T
	To   T
}

const (
	unvisited = 0
	visiting  = 1
	visited   = 2
)

type Digraph[T ID] struct {
	edges       []Edge[T]
	edgeIndex   map[Edge[T]]struct{}
	vertices    []*Vertex[T]
	vertexIndex map[T]*Vertex[T]
}

func NewDigraph[T ID]() *Digraph[T] {
	return &Digraph[T]{
		edgeIndex:   map[Edge[T]]struct{}{},
		vertexIndex: map[T]*Vertex[T]{},
	}
}

func (graph *Digraph[T]) CloneDeep() *Digraph[T] {
	cloned := &Digraph[T]{
		edges:       slices.Clone(graph.edges),
		edgeIndex:   maps.Clone(graph.edgeIndex),
		vertices:    make([]*Vertex[T], len(graph.vertices)),
		vertexIndex: make(map[T]*Vertex[T], len(graph.vertexIndex)),
	}
	for idx, vertex := range graph.vertices {
		clonedVertex := vertex.CloneDeep()
		cloned.vertices[idx] = &clonedVertex
		cloned.vertexIndex[clonedVertex.ID] = &clonedVertex
	}
	return cloned
}

func (graph *Digraph[T]) Add(id T) error {
	graph.ensureVertex(id)
	return nil
}

func (graph *Digraph[T]) Connect(from, to T) error {
	edge := Edge[T]{from, to}
	if _, ok := graph.edgeIndex[edge]; ok {
		return nil
	}

	fromVertex := graph.ensureVertex(from)
	toVertex := graph.ensureVertex(to)

	fromVertex.out.Add(to)
	toVertex.in.Add(from)

	graph.edges = append(graph.edges, edge)
	graph.edgeIndex[edge] = struct{}{}

	return nil
}

func (graph *Digraph[T]) Disconnect(from, to T) error {
	edge := Edge[T]{from, to}
	if _, ok := graph.edgeIndex[edge]; !ok {
		return nil
	}

	fromVertex := graph.vertexIndex[from]
	if fromVertex == nil {
		return nil
	}

	toVertex := graph.vertexIndex[to]
	if toVertex == nil {
		return nil
	}

	graph.removeEdge(edge)

	fromVertex.out.Remove(to)
	if fromVertex.Empty() {
		graph.removeVertex(fromVertex)
	}

	toVertex.in.Remove(from)
	if toVertex.Empty() {
		graph.removeVertex(toVertex)
	}

	return nil
}

func (graph *Digraph[T]) Remove(item T) error {
	vertex := graph.vertexIndex[item]
	if vertex == nil {
		return nil
	}
	graph.removeVertex(vertex)

	for _, out := range vertex.out.order {
		graph.removeEdge(Edge[T]{item, out})
	}
	for _, in := range vertex.in.order {
		graph.removeEdge(Edge[T]{in, item})
	}

	return nil
}

func (graph *Digraph[T]) Subgraph(ids []T) *Digraph[T] {
	sub := &Digraph[T]{
		edgeIndex:   map[Edge[T]]struct{}{},
		vertices:    make([]*Vertex[T], 0, len(ids)),
		vertexIndex: make(map[T]*Vertex[T], len(ids)),
	}

	idIndex := make(map[T]struct{}, len(ids))
	for _, id := range ids {
		idIndex[id] = struct{}{}
		sub.Add(id)
	}

	for _, id := range ids {
		v := graph.vertexIndex[id]
		if v == nil {
			continue
		}
		for _, out := range v.out.order {
			if _, ok := idIndex[out]; ok {
				sub.Connect(id, out)
			}
		}
	}

	return sub
}

func (graph *Digraph[T]) StronglyConnected() [][]T {
	var scc = newSCCState(graph)
	for _, v := range graph.vertices {
		if _, ok := scc.vertexState[v.ID]; !ok {
			scc.strongConnect(v.ID)
		}
	}
	return scc.sccs
}

func (graph *Digraph[T]) Cycles() [][]T {
	var cycles [][]T
	sccs := graph.StronglyConnected()
	for _, scc := range sccs {
		if len(scc) > 1 {
			cycles = append(cycles, scc)
		}
	}
	return cycles
}

func (graph *Digraph[T]) HasCycles() bool {
	visitStatus := make(map[T]int, len(graph.vertices))

	var walk func(item T) bool
	walk = func(item T) bool {
		if visitStatus[item] == visited {
			return false
		} else if visitStatus[item] == visiting {
			return true
		}
		visitStatus[item] = visiting
		for _, dep := range graph.vertexIndex[item].out.order {
			if walk(dep) {
				return true
			}
		}
		visitStatus[item] = visited
		return false
	}

	for _, v := range graph.vertices {
		if walk(v.ID) {
			return true
		}
	}

	return false
}

// Creates a DAG from the Digraph and returns the edges cut, using the greedy
// approximation described in Eades, Lin, Smyth: "A fast and effective heuristic for the
// feedback arc set problem".
// https://doi.org/10.1016/0020-0190(93)90079-O
//
// This has not been optimised at _all_ and is very slow with more than 10k edges.
func (graph *Digraph[T]) DAGEades() (dag *Digraph[T], cut []Edge[T]) {
	dag = graph.CloneDeep()
	var s1, s2 []T

	for len(dag.vertices) > 0 {

	sinkLoop:
		for len(dag.vertices) > 0 {
			for _, vertex := range dag.vertices {
				if vertex.Outdegree() == 0 {
					dag.Remove(vertex.ID)
					s2 = append(s2, vertex.ID) // Should prepend, but we reverse later
					continue sinkLoop
				}
			}
			break sinkLoop
		}

	sourceLoop:
		for len(dag.vertices) > 0 {
			for _, vertex := range dag.vertices {
				if vertex.Indegree() == 0 {
					dag.Remove(vertex.ID)
					s1 = append(s1, vertex.ID)
					continue sourceLoop
				}
			}
			break sourceLoop
		}

		if len(dag.vertices) > 0 {
			sort.Slice(dag.vertices, func(i, j int) bool {
				vi, vj := dag.vertices[i], dag.vertices[j]
				return (vi.Outdegree() - vi.Indegree()) > (vj.Outdegree() - vj.Indegree())
			})
			vertex := dag.vertices[0]
			dag.Remove(vertex.ID)
			s1 = append(s1, vertex.ID)
		}
	}

	slices.Reverse(s2)

	order := append(s1, s2...)

	for _, edge := range graph.edges {
		var isCut bool
		if edge.From == edge.To {
			isCut = true
		} else {
			fromIdx := slices.Index(order, edge.From)
			toIdx := slices.Index(order, edge.To)
			if fromIdx < 0 || toIdx < 0 {
				panic("edge not found in ordering")
			}
			isCut = fromIdx > toIdx
		}

		if isCut {
			cut = append(cut, edge)
		} else {
			dag.Connect(edge.From, edge.To)
		}
	}

	if dag.HasCycles() {
		panic("dag still has cycles")
	}

	return dag, cut
}

func (graph *Digraph[T]) ensureVertex(id T) *Vertex[T] {
	v := graph.vertexIndex[id]
	if v != nil {
		return v
	}

	v = &Vertex[T]{
		ID:  id,
		in:  newInsertOrderedSet[T](),
		out: newInsertOrderedSet[T](),
	}
	graph.vertices = append(graph.vertices, v)
	graph.vertexIndex[id] = v
	return v
}

// Low-level function, should only be called if you're mopping up vertices yourself.
func (graph *Digraph[T]) removeEdge(edge Edge[T]) {
	delete(graph.edgeIndex, edge)
	idx := slices.Index(graph.edges, edge)
	if idx >= 0 {
		graph.edges = slices.Delete(graph.edges, idx, idx+1)
	}
}

// Low-level function, should only be called if you're mopping up edges yourself.
func (graph *Digraph[T]) removeVertex(vertex *Vertex[T]) {
	delete(graph.vertexIndex, vertex.ID)
	idx := slices.Index(graph.vertices, vertex)
	if idx >= 0 {
		graph.vertices = slices.Delete(graph.vertices, idx, idx+1)
	}
}
