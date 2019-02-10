package graph

import (
	"testing"
)

func TestSort(t *testing.T) {
	WithLogger(t.Log)
	g := New(5)
	g.AddEdge(0, 1)
	g.AddEdge(0, 2)
	g.AddEdge(0, 3)
	g.AddEdge(1, 4)
	g.AddEdge(2, 4)
	g.AddEdge(3, 4)

	sorted := Sort(g)

	t.Logf("sorted list: %v\n", sorted)

	if len(sorted) != g.Vertices() {
		t.Fatal("the graph is not a dag!")
	}
}

func equals(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}

	for i, elt := range a {
		if elt != b[i] {
			return false
		}
	}

	return true
}

func TestDFS(t *testing.T) {
	g := New(4)
	g.AddEdge(0, 1)
	g.AddEdge(0, 2)
	g.AddEdge(1, 2)

	order := []int{}
	DFS(g, func(w int) error {
		order = append(order, w)
		return nil
	})

	t.Logf("order=%v", order)

	if len(order) != g.Vertices() {
		t.Fatal("expected dfs order to match vertices")
	}

}
