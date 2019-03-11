package graph

import (
	"reflect"
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

	if after(sorted, 0, 1, 2, 3) {
		t.Fatalf("the graph is not sorted correctly, sorted: %v", sorted)
	}
	if after(sorted, 1, 4) {
		t.Fatalf("the graph is not sorted correctly, sorted: %v", sorted)
	}
	if after(sorted, 2, 4) {
		t.Fatalf("the graph is not sorted correctly, sorted: %v", sorted)
	}
	if after(sorted, 3, 4) {
		t.Fatalf("the graph is not sorted correctly, sorted: %v", sorted)
	}
}

func TestSort2(t *testing.T) {
	WithLogger(t.Log)

	g := newGraph(3)
	g.addEdge(0, 1)

	sorted := sort(g)

	if after(sorted, 0, 1) {
		t.Fatalf("0 is a parent of 1, sorted = %v", sorted)
	}
}

func indexOf(g []int, a int) int {
	for i := range g {
		if a == g[i] {
			return i
		}
	}
	return -1
}

func after(g []int, a int, others ...int) bool {
	return !before(g, a, others...)
}

func before(g []int, a int, others ...int) bool {
	indexA := indexOf(g, a)
	for _, b := range others {
		indexB := indexOf(g, b)
		if indexA > indexB {
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

	if !reflect.DeepEqual(order, []int{2, 1, 0, 3}) {
		t.Fatal("dfs did not work")
	}
}
