package graph

import (
	"testing"
)

func TestSort(t *testing.T) {
	WithLogger(t.Log)
	g := newGraph(5)
	g.addEdge(2, 4)
	g.addEdge(0, 1)
	g.addEdge(0, 2)
	g.addEdge(1, 4)
	g.addEdge(3, 4)
	g.addEdge(0, 3)

	sorted := sort(g)

	t.Logf("sorted list: %v\n", sorted)

	if sorted[0] != 0 {
		t.Fatalf("sorted=%v, expected 0 to be in first position", sorted)
	}
	if sorted[4] != 4 {
		t.Fatalf("sorted=%v, expected 4 to be in last position", sorted)
	}
	if len(sorted) != g.vertices() {
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
	g := newGraph(4)
	g.addEdge(0, 1)
	g.addEdge(0, 2)
	g.addEdge(1, 2)

	order := []int{}
	dfs(g, func(w int) error {
		order = append(order, w)
		return nil
	})

	t.Logf("order=%v", order)

	if len(order) != g.vertices() {
		t.Fatal("expected dfs order to match vertices")
	}

}
