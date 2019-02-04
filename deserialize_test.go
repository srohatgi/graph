package graph

import "testing"

func TestBuildGraph(t *testing.T) {

	resources := []*Resource{{
		Name: "mykin",
		Type: "kinesis",
	}, {
		Name: "mydyn",
		Type: "dynamo",
	}, {
		Name:      "mydep1",
		Type:      "deployment",
		DependsOn: []string{"mykin"},
	}}

	g := buildGraph(resources)

	t.Logf("graph = %v\n", g)

	if g.Vertices() != 3 {
		t.Fatal("vertices incorrect")
	}

	if len(g.Adjascent(0)) != 1 || g.Adjascent(0)[0] != 2 {
		t.Fatal("incorrect edges")
	}
}
