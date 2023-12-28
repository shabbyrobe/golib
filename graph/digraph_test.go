package graph

import (
	"fmt"
	"math/rand"
	"reflect"
	"testing"
)

func TestDigraphStronglyConnected(t *testing.T) {
	g := NewDigraph[string]()
	g.Connect("a", "b")
	g.Connect("a", "c")
	g.Connect("c", "b")
	g.Connect("b", "a")
	g.Connect("c", "a")

	g.Connect("d", "e")
	g.Connect("e", "d")

	expected := [][]string{
		{"c", "b", "a"},
		{"e", "d"},
	}
	if !reflect.DeepEqual(g.StronglyConnected(), expected) {
		t.Fatal()
	}
}

func TestDigraphDAG(t *testing.T) {
	g := NewDigraph[string]()
	g.Connect("z", "a")
	g.Connect("a", "b")
	g.Connect("a", "c")
	g.Connect("c", "b")
	g.Connect("b", "a")
	g.Connect("c", "a")

	g.Connect("d", "e")
	g.Connect("e", "d")
	g.Connect("e", "f")
	g.Connect("f", "f")

	dag, cut := g.DAGEades()
	expectedEdges := []Edge[string]{
		{"z", "a"},
		{"a", "b"},
		{"c", "b"},
		{"c", "a"},
		{"e", "d"},
		{"e", "f"},
	}
	expectedCut := [][]string{
		{"c", "b", "a"},
		{"e", "d"},
	}
	if !reflect.DeepEqual(dag.edges, expectedEdges) {
		t.Fatalf("%+v != %+v", dag.edges, expectedEdges)
	}
	if !reflect.DeepEqual(dag.edges, expectedEdges) {
		t.Fatalf("%+v != %+v", cut, expectedCut)
	}
}

func TestDigraphDAGSpam(t *testing.T) {
	type testCase struct {
		edges  int
		maxval int64
		seed   int64
	}

	var cases = []testCase{}

	for seed := int64(0); seed < 1000; seed++ {
		cases = append(cases,
			testCase{100, 10, seed},
			testCase{100, 20, seed},
			testCase{100, 50, seed},
			testCase{100, 100, seed},
		)
	}
	for seed := int64(0); seed < 5; seed++ {
		cases = append(cases,
			testCase{1000, 100, seed},
			testCase{1000, 500, seed},
			testCase{1000, 2000, seed},
			testCase{10000, 5000, seed},
		)
	}

	withoutCycles := 0
	total := 0

	for idx, tc := range cases {
		t.Run(fmt.Sprintf("%d/%d/%d/%d", idx, tc.edges, tc.maxval, tc.seed), func(t *testing.T) {
			total++
			g := NewDigraph[int64]()
			rng := rand.New(rand.NewSource(tc.seed))
			for i := 0; i < tc.edges; i++ {
				g.Connect(rng.Int63n(tc.maxval), rng.Int63n(tc.maxval))
			}
			if !g.HasCycles() {
				withoutCycles++
			}

			dag, _ := g.DAGEades()
			if dag.HasCycles() {
				t.Fatal()
			}
			if len(dag.Cycles()) > 0 {
				t.Fatal()
			}
			if len(dag.StronglyConnected()) != len(dag.vertices) {
				t.Fatal()
			}
		})
	}

	if float64(withoutCycles)/float64(total) > 0.1 {
		panic("too many spam tests don't have cycles")
	}
}

var BenchDigraphInt64 *Digraph[int64]

func BenchmarkDigraphDAG(b *testing.B) {
	var edges int64 = 10000
	var maxVal int64 = 1_000_000
	items := make([][2]int64, edges)
	rng := rand.New(rand.NewSource(1))
	for idx := range items {
		items[idx] = [2]int64{rng.Int63n(maxVal), rng.Int63n(maxVal)}
	}

	g := NewDigraph[int64]()
	for _, i := range items {
		g.Connect(i[0], i[1])
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		BenchDigraphInt64, _ = g.DAGEades()
	}
}

// func TestDigraphDAGSSCC(t *testing.T) {
//     g := NewDigraph[int64]()
//     rng := rand.New(rand.NewSource(0))
//     for i := 0; i < 200; i++ {
//         g.Connect(rng.Int63n(50), rng.Int63n(50))
//     }
//
//     for idx, scc := range g.StronglyConnected() {
//         sub := g.Subgraph(scc)
//         dot, _ := Dot(sub)
//         os.WriteFile(fmt.Sprintf("/tmp/%d.dot", idx), []byte(dot), 0600)
//     }
// }
