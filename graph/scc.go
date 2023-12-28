package graph

type sccVertexState struct {
	index   int
	lowLink int
	onStack bool
}

type sccState[T ID] struct {
	graph       *Digraph[T]
	nextIndex   int
	vertexState map[T]*sccVertexState
	stack       []T
	sccs        [][]T
}

func (state *sccState[T]) reset(graph *Digraph[T]) {
	state.graph = graph
	state.nextIndex = 0

	clear(state.vertexState)
	clear(state.stack)
	state.stack = state.stack[:0]

	clear(state.sccs)
	state.sccs = state.sccs[:0]
}

func newSCCState[T ID](graph *Digraph[T]) *sccState[T] {
	return &sccState[T]{
		vertexState: map[T]*sccVertexState{},
		graph:       graph,
	}
}

// Strongly connect components using Tarjan's algorithm, so we can detect cycles
// in the graph.
func (scc *sccState[T]) strongConnect(id T) {
	// Set the depth index for v to the smallest unused index
	scc.vertexState[id] = &sccVertexState{
		index:   scc.nextIndex,
		lowLink: scc.nextIndex,
		onStack: true,
	}
	scc.stack = append(scc.stack, id)
	scc.nextIndex++

	// Consider successors of v
	for _, edge := range scc.graph.edges {
		if edge.From != id {
			continue
		}
		if toState, ok := scc.vertexState[edge.To]; !ok {
			// Successor w has not yet been visited; recurse on it
			scc.strongConnect(edge.To)

			fromState := scc.vertexState[edge.From]
			toState := scc.vertexState[edge.To]

			// v.lowlink := min(v.lowlink, w.lowlink)
			if toState.lowLink < fromState.lowLink {
				fromState.lowLink = toState.lowLink
			}

		} else if toState.onStack {
			// Successor w is in stack S and hence in the current SCC
			// If w is not on stack, then (v, w) is an edge pointing to an SCC already found and must be ignored
			// Note: The next line may look odd - but is correct.
			// It says w.index not w.lowlink; that is deliberate and from the original paper
			// v.lowlink := min(v.lowlink, w.index)
			fromState := scc.vertexState[edge.From]
			if toState.index < fromState.lowLink {
				fromState.lowLink = toState.index
			}
		}
	}

	// If v is a root node, pop the stack and generate an SCC
	if state := scc.vertexState[id]; state.lowLink == state.index {
		var component []T
		var w T

		// Required so we can distinguish between the first iteration and a vertex ID
		// that just happens to be the zero value:
		var wset bool

		for !wset || w != id {
			w, scc.stack = scc.stack[len(scc.stack)-1], scc.stack[:len(scc.stack)-1]
			wset = true
			wState := scc.vertexState[w]
			wState.onStack = false
			component = append(component, w)
		}
		scc.sccs = append(scc.sccs, component)
	}
}
