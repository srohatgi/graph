package graph

import (
	"strings"
	"testing"
)

// TestBasic simple stuff
func TestBasic(t *testing.T) {
	g := newGraph(2)

	g.addEdge(0, 1)

	if len(g.adjascent(0)) != 1 {
		t.Fatal("expected 1 adjascent vertex")
	}
}

// TestFromReader using a textual representation
func TestFromReader(t *testing.T) {
	serialized := `3
3
0 1
1 2
0 2
`
	r := strings.NewReader(serialized)
	g, err := newFromReader(r)

	if err != nil {
		t.Fatal(err)
	}

	if g.vertices() != 3 {
		t.Fatalf("incorrect number of vertices expected 2, got %v", g.vertices())
	}

	if len(g.adjascent(0)) != 2 {
		t.Fatal("expected 0 vertex to be adjascent to 1 and 2")
	}

	t.Logf("graph = %v\n", g)
}
