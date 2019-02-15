package graph

import "errors"

// BuildCache allows a loose form of communication
type BuildCache map[string]interface{}

// Factory allows specialized builder creation
type Factory interface {
	Create(resource *Resource) Builder
}

// Builder allows deletion or update of resources
type Builder interface {
	Get() *Resource
	Delete() error
	Update(in []Property) ([]Property, error)
}

// Sync up all resources
func Sync(resources []*Resource, toDelete bool, factory Factory) error {
	g := buildGraph(resources)

	builders := []Builder{}

	for _, r := range resources {
		builders = append(builders, factory.Create(r))
	}

	if toDelete {
		return deleteSync(builders, g)
	}

	return createSync(builders, g)
}

func createSync(builders []Builder, g *Graph) error {
	ordered := Sort(g)

	buildCache := map[string][]Property{}
	var err error

	resourcesLeft := len(ordered)
	maxAttempts := len(ordered)

	for maxAttempts > 0 && resourcesLeft > 0 && err == nil {
		maxAttempts--
		for _, i := range ordered {
			var in, out []Property
			res := builders[i].Get()
			notReady := false
			for _, dep := range res.DependsOn {
				if _, found := buildCache[dep.ResourceName]; !found {
					// cannot proceed as this resource cannot be processed
					notReady = true
					break
				}
				in = append(in, buildCache[dep.ResourceName]...)
			}
			if notReady {
				break
			}
			out, err = builders[i].Update(in)
			if err != nil {
				break
			}
			buildCache[res.Name] = append(buildCache[res.Name], out...)
			resourcesLeft--
		}
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return err
}

func reverse(in []int) {
	for i := len(in)/2 - 1; i >= 0; i-- {
		opp := len(in) - 1 - i
		in[i], in[opp] = in[opp], in[i]
	}
}

func deleteSync(builders []Builder, g *Graph) error {
	order := Sort(g)
	reverse(order)

	var err error

	for _, i := range order {
		err = builders[i].Delete()
		if err != nil {
			break
		}
	}

	return err
}
