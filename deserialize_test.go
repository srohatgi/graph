package graph

import "testing"

func TestBuildGraph(t *testing.T) {

	kinName := "mykin"

	resources := []*Resource{{
		Name: kinName,
		Type: "kinesis",
	}, {
		Name: "mydyn",
		Type: "dynamo",
	}, {
		Name:      "mydep1",
		Type:      "deployment",
		DependsOn: []string{kinName},
	}}

	g := buildGraph(resources)

	t.Logf("graph = %v\n", g)

	if g.vertices() != 3 {
		t.Fatal("vertices incorrect")
	}

	if len(g.adjascent(0)) != 1 || g.adjascent(0)[0] != 2 {
		t.Fatal("incorrect edges")
	}
}
