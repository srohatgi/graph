package graph

import "errors"

// BuildCache allows a loose form of communication
type BuildCache map[string]string

// Builder allows deletion or update of things
type Builder interface {
	Delete(cache BuildCache) error
	Update(cache BuildCache) error
}

// Kinesis is aws data stream
type Kinesis struct {
	*Resource
}

func (k *Kinesis) Update(cache BuildCache) error { return nil }
func (k *Kinesis) Delete(cache BuildCache) error { return nil }

// Dynamo is aws data table
type Dynamo struct {
	*Resource
}

func (k *Dynamo) Update(cache BuildCache) error { return nil }
func (k *Dynamo) Delete(cache BuildCache) error { return nil }

// Deployment is kubernetes deployment
type Deployment struct {
	*Resource
}

func (k *Deployment) Update(cache BuildCache) error { return nil }
func (k *Deployment) Delete(cache BuildCache) error { return nil }

// Convert enables a resource to be buildable
func Convert(resource *Resource) Builder {
	var builder Builder
	switch resource.Type {
	case "kinesis":
		builder = &Kinesis{resource}
	case "dynamo":
		builder = &Dynamo{resource}
	case "deployment":
		builder = &Deployment{resource}
	}

	return builder
}

// Sync up all resources
func Sync(resources []*Resource, toDelete bool) error {
	g := BuildGraph(resources)
	ordered := Sort(g)

	buildCache := map[string]string{}
	var err error
	for _, i := range ordered {
		builder := Convert(resources[i])
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
