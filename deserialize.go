package graph

// Resource models a virtual service
type Resource struct {
	// Name is expected to be unique in a given slice of Resource's
	Name string
	// Type is expected to be used for creating Builder's
	Type string
	// Bag is a convenient pointer, unused by the library
	Bag interface{}
	// Properties are input to the Builder
	Properties []Property
	// DependsOn are dependencies used for sorting Resource slice
	DependsOn []Dependency
}

// Property is an arbitrary name value pair
type Property struct {
	// Name is unique per Resource
	Name  string
	Value string
}

// Dependency captures resource dependencies
type Dependency struct {
	// ResourceName refers to the unique Name in a slice of Resource's
	ResourceName string
	// Properties used by the Builder
	Properties []Property
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
