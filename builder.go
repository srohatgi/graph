// Package graph library may be useful for developers tasked with storage, compute and
// network management for cloud microservices.  The methods and types in the
// library enforce an extensible and modular programming model. The library
// assumes that a declarative model of defining a resource exists.
//
// Types and Values
//
// A Resource is a declarative abstraction of a storage, compute or network
// service. A Builder is an interface for managing a Resource. A Factory is
// an interface for creating a Builder, given a Resource.
//
// The library manages a collection of related resources at a given time. Hence,
// multiple errors may be produced concurrently. There is a handy ErrorMap struct
// that allows developers to parse out different errors and map them to individual
// resources. Use the following code snippet:
//  if em, ok := err.(*ErrorMap); ok {
//    for resourceIndex, err := range *em {
//      fmt.Printf("resource %d creation had error %v\n", resourceIndex, err)
//    }
//  }
//
// Methods
//
// The library provides ways to manage an ordered collection of resources. Order is
// specified by naming resources uniquely in a collection, and having a resource
// depend on a set of other resources within the same collection.
//
// The Sync() method provides a method for managing a collection of resources:
//  err := Sync(resources, false, factory) // refer to signature below
package graph

import (
	"errors"
	"fmt"
	"sync"
)

// buildCache allows a loose form of communication
type buildCache map[string]interface{}

// Factory allows specialized Builder creation.
type Factory interface {
	// Create produces a new Builder. A Factory instance may inject
	// context.Context object to be used inside of the Builder interface
	// methods.
	// NOTE: One Builder instance maps to a single Resource instance.
	Create(resource *Resource) Builder
}

// Builder methods enable Resource management.
type Builder interface {
	// Get retrieves underlying Resource instance.
	Get() *Resource
	// Delete the Resource.
	Delete() error
	// Update or if not existing, create the Resource.
	Update(in []Property) ([]Property, error)
}

type builderOutput struct {
	result error
	out    []Property
}

type protoBuilder struct {
	r     *Resource
	udef  interface{}
	updFn func(interface{}, []Property) ([]Property, error)
	delFn func(interface{}) error
}

func (p *protoBuilder) Get() *Resource                           { return p.r }
func (p *protoBuilder) Update(in []Property) ([]Property, error) { return p.updFn(p.udef, in) }
func (p *protoBuilder) Delete() error                            { return p.delFn(p.udef) }

// MakeBuilder is a convenient utility to create Builder's in a cheap way.
// NOTE: uDef is a custom generic struct that is injected into updFn & delFn
func MakeBuilder(r *Resource, uDef interface{}, updFn func(interface{}, []Property) ([]Property, error), delFn func(interface{}) error) Builder {
	return &protoBuilder{r, uDef, updFn, delFn}
}

// Sync method is used to enforce the programming model. Internally, the method
// maps the Resource slice to a Builder slice (using the Factory instance), and
// then executes appropriate Builder interface methods. When a subset of resources
// can be updated or created in parallel, the method attempts to do it.
func Sync(resources []*Resource, toDelete bool, factory Factory) error {
	g := buildGraph(resources)

	logger("starting sync")

	builders := []Builder{}

	for _, r := range resources {
		b := factory.Create(r)
		if b == nil {
			return fmt.Errorf("unable to create builder for resource: %v", *r)
		}
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

		errs := ErrorMap{}
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

	logger("order of deletion", order)

	var err error

	for _, i := range order {
		err = builders[i].Delete()
		if err != nil {
			break
		}
	}

	return err
}
