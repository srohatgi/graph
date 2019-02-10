package graph

// Property is an arbitrary name value pair
type Property struct {
	ResourceName *string
	Name         string
	Value        string
}

// Resource models a virtual service
type Resource struct {
	Name       string
	Type       string
	Properties []Property
}

func buildGraph(resources []*Resource) *Graph {
	parents := map[int][]int{}
	indexes := map[string]int{}

	for i := range resources {
		indexes[resources[i].Name] = i
	}

	for i := range resources {
		for _, prop := range resources[i].Properties {
			if prop.ResourceName != nil && *prop.ResourceName != resources[i].Name {
				parents[i] = append(parents[i], indexes[*prop.ResourceName])
			}
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
