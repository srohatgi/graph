package graph

import (
	"errors"
	"sync"
)

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

	logger("starting sync")

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
		execList := []int{}
		for _, i := range ordered {
			res := builders[i].Get()
			ready := true
			for _, dep := range res.DependsOn {
				if _, found := buildCache[dep.ResourceName]; !found {
					// cannot proceed as this resource cannot be processed
					ready = false
					break
				}
			}
			if !ready {
				break
			}
			execList = append(execList, i)
		}

		// execute nodes that are ready
		var wg sync.WaitGroup
		errs := map[int]chan error{}

		for _, i := range execList {
			wg.Add(1)
			errs[i] = make(chan error, 1)
			go func(b Builder, c chan error) {
				defer wg.Done()
				c <- execute(b, buildCache)
			}(builders[i], errs[i])
		}

		wg.Wait()
		logger("done waiting")

		errCnt := 0
		for i, c := range errs {
			e := <-c
			if e != nil {
				logger("error executing builder", "builder", builders[i], "error", e)
				err = e
				errCnt++
			}
		}

		resourcesLeft -= len(execList) - errCnt
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return err
}

func execute(b Builder, cache map[string][]Property) error {
	var in []Property
	res := b.Get()
	for _, dep := range res.DependsOn {
		in = append(in, cache[dep.ResourceName]...)
	}

	_, err := b.Update(in)
	if err != nil {
		return err
	}
	//cache[res.Name] = append(cache[res.Name], out...)
	return nil
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
