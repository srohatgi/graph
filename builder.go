package graph

import "errors"

// BuildCache allows a loose form of communication
type BuildCache map[string]interface{}

// Factory allows specialized builder creation
type Factory interface {
	Create(resource *Resource) Builder
}

// Builder allows deletion or update of things
type Builder interface {
	Delete(cache BuildCache) error
	Update(cache BuildCache) error
}

// Sync up all resources
func Sync(resources []*Resource, toDelete bool, factory Factory) error {
	g := buildGraph(resources)
	ordered := Sort(g)

	buildCache := map[string]interface{}{}
	var err error
	for _, i := range ordered {
		builder := factory.Create(resources[i])
		if builder == nil {
			err = errors.New("unknown resource type: " + resources[i].Type)
			break
		}
		if toDelete {
			err = builder.Delete(buildCache)
		} else {
			err = builder.Update(buildCache)
		}
		if err != nil {
			break
		}
	}

	return err
}
