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
	"sync"
)

// Dependency captures inter resource dependencies
type Dependency struct {
	FromResourceName string
	FieldName        string
	ToFieldName      string
}

// Resource is an abstract declarative definition for compute, storage and network services.
// Examples: AWS Kinesis, AWS CloudFormation, Kubernetes Deployment etc.
type Resource interface {
	// Get retrieves underlying Resource instance name. This allows creation
	// of multiple resources of the same Type.
	Name() string
	// Get retrieves underlying Resource type.
	Type() string
	// Dependencies fetches a given Builder's dependency list.
	Dependencies() []Dependency
	// Delete the Resource.
	Delete() error
	// Update or if not existing, create the Resource.
	Update() (string, error)
}

type builderOutput struct {
	result error
	status string
}

type protoBuilder struct {
	name         string
	resourceType string
	dependencies []Dependency
	uDef         interface{}
	updFn        func(interface{}) (string, error)
	delFn        func(interface{}) error
}

func (p *protoBuilder) Name() string               { return p.name }
func (p *protoBuilder) Type() string               { return p.resourceType }
func (p *protoBuilder) Update() (string, error)    { return p.updFn(p.uDef) }
func (p *protoBuilder) Delete() error              { return p.delFn(p.uDef) }
func (p *protoBuilder) Dependencies() []Dependency { return p.dependencies }

// MakeResource is a convenient utility to create Resource's in a cheap way.
// NOTE: uDef is a custom generic struct that is injected into updFn & delFn
func MakeResource(name, resourceType string, dependencies []Dependency, uDef interface{}, updFn func(interface{}) (string, error), delFn func(interface{}) error) Resource {
	return &protoBuilder{name, resourceType, dependencies, uDef, updFn, delFn}
}

// Sync method is used to enforce the programming model. Internally, the method
// maps the Resource slice to a Builder slice (using the Factory instance), and
// then executes appropriate Builder interface methods. When a subset of resources
// can be updated or created in parallel, the method attempts to do it.
func Sync(resources []Resource, toDelete bool) error {
	g := buildGraph(resources)

	logger("starting sync")

	if toDelete {
		return deleteSync(resources, g)
	}

	return createSync(resources, g)
}

func getProperty(r Resource, name string) interface{} {
	return nil
}

func setProperty(r Resource, name string, value interface{}) {

}

func createSync(resources []Resource, g *graph) error {
	ordered := sort(g)

	var err error

	resourcesLeft := len(ordered)
	maxAttempts := len(ordered)

	buildCache := map[string]bool{}
	for maxAttempts > 0 && resourcesLeft > 0 && err == nil {
		maxAttempts--
		execList := []int{}
		for _, i := range ordered {

			res := resources[i]
			// check if we've already executed
			if _, alreadyExecuted := buildCache[res.Name()]; alreadyExecuted {
				logger("already executed", i)
				continue
			}

			ready := true
			for _, dep := range res.Dependencies() {
				if _, found := buildCache[dep.FromResourceName]; !found {
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

			go func(b Resource, c chan builderOutput) {
				defer wg.Done()
				c <- execute(b, resources)
			}(resources[i], output[i])
		}

		wg.Wait()

		errs := ErrorMap{}
		for i, c := range output {
			e := <-c
			if e.result != nil {
				logger("error executing resource", "resource", resources[i], "error", e)
				errs[i] = e.result
				continue
			}

			name := resources[i].Name()
			buildCache[name] = true
		}

		resourcesLeft -= len(execList) - errs.Size()
		err = errs.Get()
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return err
}

func execute(r Resource, cache []Resource) builderOutput {
	for _, dep := range r.Dependencies() {
		for _, from := range cache {
			if from.Name() == dep.FromResourceName {
				prop := getProperty(from, dep.FieldName)
				setProperty(r, dep.ToFieldName, prop)
			}
		}
	}

	out, err := r.Update()
	if err != nil {
		return builderOutput{result: err}
	}
	return builderOutput{nil, out}
}

func reverse(in []int) {
	for i := len(in)/2 - 1; i >= 0; i-- {
		opp := len(in) - 1 - i
		in[i], in[opp] = in[opp], in[i]
	}
}

func deleteSync(resources []Resource, g *graph) error {
	order := sort(g)
	reverse(order)

	logger("order of deletion", order)

	var err error

	for _, i := range order {
		err = resources[i].Delete()
		if err != nil {
			break
		}
	}

	return err
}
