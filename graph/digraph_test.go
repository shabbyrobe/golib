package graph

import (
	"errors"
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

func TestDigraphDepths(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		g := NewDigraph[string]()
		expected := map[string]int{}
		depths, err := g.Depths()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(depths, expected) {
			t.Fatal(depths, expected)
		}
	})

	t.Run("cycle1", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("yep", "yep")
		_, err := g.Depths()
		if err == nil {
			t.Fatal()
		}
	})

	t.Run("cycle2", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("yep", "ahoy")
		g.Connect("ahoy", "yep")
		_, err := g.Depths()
		if err == nil {
			t.Fatal()
		}
	})

	t.Run("cycle2-with-source", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("source", "yep")
		g.Connect("yep", "ahoy")
		g.Connect("ahoy", "yep")
		_, err := g.Depths()
		if err == nil {
			t.Fatal()
		}
	})

	t.Run("same-tree-different-depths", func(t *testing.T) {
		g := NewDigraph[string]()

		g.Connect("p1", "a")
		g.Connect("p2", "p3")
		g.Connect("p3", "a")

		g.Connect("a", "b")
		g.Connect("b", "c")
		g.Connect("a", "d")

		expected := map[string]int{
			"p1": 2,
			"p2": 1,
			"p3": 2,
			"a":  3,
			"b":  4,
			"c":  5,
			"d":  5,
		}
		depths, err := g.Depths()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(depths, expected) {
			t.Fatal(depths, expected)
		}
	})

	t.Run("nodes", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("a", "b")
		g.Connect("a", "c")
		g.Connect("c", "d")
		g.Connect("b", "d")
		g.Connect("d", "f")
		g.Connect("e", "f")
		g.Connect("a", "e")

		expected := map[string]int{
			"a": 1,
			"b": 2,
			"c": 2,
			"d": 3,
			"e": 3,
			"f": 4,
		}
		depths, err := g.Depths()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(depths, expected) {
			t.Fatal(depths, expected)
		}
	})

	t.Run("multiple-sources", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("a", "b")
		g.Connect("c", "d")
		g.Connect("b", "d")

		expected := map[string]int{
			"a": 1,
			"b": 2,
			"c": 2,
			"d": 3,
		}
		depths, err := g.Depths()
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(depths, expected) {
			t.Fatal(depths, expected)
		}
	})
}

func TestDigraphSortedTopological(t *testing.T) {
	t.Run("dag", func(t *testing.T) {
		g := NewDigraph[string]()

		g.Connect("z", "a")
		g.Connect("a", "b")
		g.Connect("a", "c")
		g.Connect("a", "d")
		g.Connect("c", "d")
		g.Connect("c", "b")

		sorted, err := g.SortedTopological()
		if err != nil {
			t.Fatal(err)
		}

		expected := []string{"b", "d", "c", "a", "z"}
		if !reflect.DeepEqual(sorted, expected) {
			t.Fatalf("%+v != %+v", sorted, expected)
		}
	})

	t.Run("one-node", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Add("z")

		sorted, err := g.SortedTopological()
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"z"}
		if !reflect.DeepEqual(sorted, expected) {
			t.Fatalf("%+v != %+v", sorted, expected)
		}
	})

	t.Run("one-edge", func(t *testing.T) {
		g := NewDigraph[string]()
		g.Connect("a", "b")

		sorted, err := g.SortedTopological()
		if err != nil {
			t.Fatal(err)
		}
		expected := []string{"b", "a"}
		if !reflect.DeepEqual(sorted, expected) {
			t.Fatalf("%+v != %+v", sorted, expected)
		}
	})

	t.Run("cycle", func(t *testing.T) {
		g := NewDigraph[string]()

		g.Connect("a", "b")
		g.Connect("a", "c")
		g.Connect("a", "d")
		g.Connect("c", "a")

		_, err := g.SortedTopological()
		if err == nil {
			t.Fatal()
		} else if !errors.Is(err, ErrCycleDetected) {
			t.Fatal(err)
		}
	})
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
