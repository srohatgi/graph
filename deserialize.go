package graph

// Resource models a virtual service
type Resource struct {
	// unique name of a resource in a given slice of Resource's
	Name       string
	Type       string
	Bag        interface{}
	Properties []Property
	DependsOn  []Dependency
}

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

func buildGraph(resources []*Resource) *graph {
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

	g := newGraph(len(resources))

	for w, arr := range parents {
		for _, v := range arr {
			g.addEdge(v, w)
		}
	}

	return g
}
