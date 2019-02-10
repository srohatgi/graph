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
	Delete() error
	Update(properties []Property) ([]Property, error)
}

// Sync up all resources
func Sync(resources []*Resource, toDelete bool, factory Factory) error {
	g := buildGraph(resources)
	ordered := Sort(g)

	buildCache := map[string][]Property{}
	var err error
	for _, i := range ordered {
		builder := factory.Create(resources[i])
		if builder == nil {
			err = errors.New("unknown resource type: " + resources[i].Type)
			break
		}
		if toDelete {
			err = builder.Delete()
		} else {
			var in, out []Property
			for _, prop := range resources[i].Properties {
				if prop.ResourceName != nil && *prop.ResourceName != resources[i].Name {
					for _, nprop := range buildCache[*prop.ResourceName] {
						if prop.Name == nprop.Name {
							in = append(in, Property{prop.ResourceName, nprop.Name, nprop.Value})
						}
					}
				}
			}
			out, err = builder.Update(in)
			if err == nil {
				buildCache[resources[i].Name] = append(buildCache[resources[i].Name], out...)
			}
		}
		if err != nil {
			break
		}
	}

	return err
}
