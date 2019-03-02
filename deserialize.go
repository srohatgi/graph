package graph

func buildGraph(resources []Resource) *graph {
	parents := map[int]map[int]bool{}
	indexes := map[string]int{}

	for i := range resources {
		indexes[resources[i].Name()] = i
	}

	for i := range resources {
		parents[i] = map[int]bool{}
		for _, dep := range resources[i].Dependencies() {
			parents[i][indexes[dep.FromResource]] = true
		}
	}

	g := newGraph(len(resources))

	for w, m := range parents {
		for k := range m {
			g.addEdge(k, w)
		}
	}

	return g
}
