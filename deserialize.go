package graph

// Property is an arbitrary name value pair
type Property struct {
	Name  string
	Value string
}

// Resource models a virtual service
type Resource struct {
	Name       string
	Type       string
	Properties []Property
	DependsOn  []string
}

func buildGraph(resources []*Resource) *Graph {
	parents := map[int][]int{}
	indexes := map[string]int{}

	for i := range resources {
		indexes[resources[i].Name] = i
	}

	for i := range resources {
		for _, dep := range resources[i].DependsOn {
			parents[i] = append(parents[i], indexes[dep])
		}
	}

	g := New(len(resources))

	for w, arr := range parents {
		for _, v := range arr {
			g.AddEdge(v, w)
		}
	}

	return g
}
