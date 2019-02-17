package graph

import (
	"errors"
	"sync"
)

// buildCache allows a loose form of communication
type buildCache map[string]interface{}

// Factory allows specialized builder creation
type Factory interface {
	// Create is an initializer for a resource type
	// NOTE: a Builder is created per Resource instance
	Create(resource *Resource) Builder
}

// Builder allows deletion or update of resources
type Builder interface {
	// Get retrieves underlying Resource instance
	Get() *Resource
	// Delete the Resource
	Delete() error
	// Update or if not existing, create the Resource
	Update(in []Property) ([]Property, error)
}

type builderOutput struct {
	result error
	out    []Property
}

// Sync maps Resource's to Builder's, and then performs either an Update()
// or Delete() operation
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

func createSync(builders []Builder, g *graph) error {
	ordered := sort(g)

	var err error

	resourcesLeft := len(ordered)
	maxAttempts := len(ordered)

	buildCache := map[string][]Property{}
	for maxAttempts > 0 && resourcesLeft > 0 && err == nil {
		maxAttempts--
		execList := []int{}
		for _, i := range ordered {

			res := builders[i].Get()
			// check if we've already executed
			if _, alreadyExecuted := buildCache[res.Name]; alreadyExecuted {
				logger("already executed", i)
				continue
			}

			ready := true
			for _, dep := range res.DependsOn {
				if _, found := buildCache[dep]; !found {
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
		output := map[int]chan builderOutput{}

		logger("executing ", execList)
		for _, i := range execList {
			wg.Add(1)
			output[i] = make(chan builderOutput, 1)

			go func(b Builder, c chan builderOutput) {
				defer wg.Done()
				c <- execute(b, buildCache)
			}(builders[i], output[i])
		}

		wg.Wait()

		var errs ErrorSlice
		for i, c := range output {
			e := <-c
			if e.result != nil {
				logger("error executing builder", "builder", builders[i], "error", e)
				errs[i] = e.result
				continue
			}

			name := builders[i].Get().Name
			buildCache[name] = append(buildCache[name], e.out...)
		}

		resourcesLeft -= len(execList) - errs.Size()
		err = errs.Get()
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return err
}

func execute(b Builder, cache map[string][]Property) builderOutput {
	var in []Property
	res := b.Get()
	for _, dep := range res.DependsOn {
		in = append(in, cache[dep]...)
	}

	out, err := b.Update(in)
	if err != nil {
		return builderOutput{err, nil}
	}
	return builderOutput{nil, out}
}

func reverse(in []int) {
	for i := len(in)/2 - 1; i >= 0; i-- {
		opp := len(in) - 1 - i
		in[i], in[opp] = in[opp], in[i]
	}
}

func deleteSync(builders []Builder, g *graph) error {
	order := sort(g)
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
