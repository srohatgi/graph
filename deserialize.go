package graph

// Resource is an abstract declarative definition for compute, storage and network services.
// Examples: AWS Kinesis, AWS CloudFormation, Kubernetes Deployment etc.
type Resource struct {
	// Name provides uniqueness for a given slice of resources.
	Name string
	// Type is a grouping of similar resources.
	Type string
	// Properties allow resources to depend on each other.
	Properties []Property
	// DependsOn enforces order of creation and deletion of resources in a given slice. Each
	// string in the slice refers to a Resource.Name
	DependsOn []string
}

// Property captures data required or produced when a given Resource is created/ updated.
// Example: ARN of AWS Kinesis stream is produced when a new Kinesis stream is created. This
// same property may be consumed in a new Kubernetes Deployment.
type Property struct {
	Name  string
	Value string
}

func buildGraph(resources []*Resource) *graph {
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

	g := newGraph(len(resources))

	for w, arr := range parents {
		for _, v := range arr {
			g.addEdge(v, w)
		}
	}

	return g
}
