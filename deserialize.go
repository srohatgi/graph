package graph

// Property is an arbitrary name value pair
type Property struct {
	Name  string
	Value string
}

// Dependency captures resource dependencies
type Dependency struct {
	ResourceName string
	Properties   []Property
}

// Resource models a virtual service
type Resource struct {
	Name       string
	Type       string
	Bag        interface{}
	Properties []Property
	DependsOn  []Dependency
}

func buildGraph(resources []*Resource) *Graph {
	parents := map[int][]int{}
	indexes := map[string]int{}

	for i := range resources {
		indexes[resources[i].Name] = i
	}

	for i := range resources {
		for _, dep := range resources[i].DependsOn {
			parents[i] = append(parents[i], indexes[dep.ResourceName])
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
