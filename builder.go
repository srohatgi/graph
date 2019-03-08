// Package graph may be useful for developers tasked with storage, compute and
// network management for cloud microservices.  The library provides types and
// functions that enable a modular, extensible programming model.
//
// We assume that a declarative model of defining and manipulating a compute or
// storage resource (example: aws cloudformation, kubernetes, etc.) will be used.
//
// Types and Values
//
// A Resource is a declarative abstraction of a storage, compute or network
// service. Resources may have a Dependency order of creation and deletion.
//
// In go, Resource is expressed as an interface with a backed by a struct having
// having public properties. Based on the dependency, these properties may be
// injected by the library at runtime.
//
// The library manages a collection of related resources at a given time. Hence,
// multiple errors may be produced concurrently. There is a handy ErrorMap struct
// that allows developers to parse out multiple errors and map them to individual
// resources.
//
// Use the following code snippet:
//  if em, ok := err.(*ErrorMap); ok {
//    for resourceIndex, err := range *em {
//      fmt.Printf("resource %d creation had error %v\n", resourceIndex, err)
//    }
//  }
//
// Functions
//
// The library manages a list of resources. Each resource may have a unique name.
//
// The Sync() function provides a method for managing a collection of resources:
//  status, err := Sync(resources, false) // refer to signature below
package graph

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// Dependency captures inter Resource dependencies. Specifying FromField
// and ToField enables copying over of value to current Resource.
type Dependency struct {
	// FromResource is another resource specified in the same slice.
	FromResource string
	// FromField is a public field from the struct implementing the Resource.
	FromField string
	// ToField is a public field in the current Resource's implementing struct.
	ToField string
}

// Resource is an abstract declarative definition for compute, storage and network services.
// Examples: AWS Kinesis, AWS CloudFormation, Kubernetes Deployment etc.
type Resource interface {
	// Get retrieves underlying Resource instance name. This allows creation
	// of multiple resources of the same Type.
	Name() string
	// Dependencies fetches a given Resource's dependency list.
	Dependencies() []Dependency
	// Delete the Resource.
	Delete() error
	// Update or if not existing, create the Resource.
	Update() (string, error)
}

type builderOutput struct {
	status string
	result error
}

type protoBuilder struct {
	ResourceName         string
	ResourceType         string
	ResourceDependencies []Dependency
	UDef                 interface{}
	UpdFn                func(interface{}) (string, error)
	DelFn                func(interface{}) error
}

func (p *protoBuilder) Name() string               { return p.ResourceName }
func (p *protoBuilder) Type() string               { return p.ResourceType }
func (p *protoBuilder) Update() (string, error)    { return p.UpdFn(p.UDef) }
func (p *protoBuilder) Delete() error              { return p.DelFn(p.DelFn) }
func (p *protoBuilder) Dependencies() []Dependency { return p.ResourceDependencies }

// MakeResource is a convenient utility to create Resource's in a cheap way.
// NOTE: uDef is a custom generic struct that is injected into updFn & delFn
func MakeResource(name, resourceType string, dependencies []Dependency, uDef interface{}, updFn func(interface{}) (string, error), delFn func(interface{}) error) Resource {
	return &protoBuilder{name, resourceType, dependencies, uDef, updFn, delFn}
}

// Sync method is used to enforce the programming model. Internally, the method
// maps the Resource slice to a Builder slice (using the Factory instance), and
// then executes appropriate Builder interface methods. When a subset of resources
// can be updated or created in parallel, the method attempts to do it.
func Sync(resources []Resource, toDelete bool) (map[string]string, error) {
	g := buildGraph(resources)

	logger("starting sync")

	if toDelete {
		return nil, deleteSync(resources, g)
	}

	return createSync(resources, g)
}

// check that resources have correct dependencies
func check(resources []Resource) error {
	cache := map[string]Resource{}

	for _, r := range resources {
		cache[r.Name()] = r
	}

	for n, r := range cache {
		if reflect.ValueOf(r).Kind() != reflect.Ptr {
			return fmt.Errorf("expected %s Resource to be implemented with a pointer to struct", n)
		}

		if reflect.ValueOf(r).Elem().Kind() != reflect.Struct {
			return fmt.Errorf("expected %s Resource to be implemented using a pointer to struct", n)
		}

		// validate each dependency
		for _, dep := range r.Dependencies() {
			if len(dep.ToField) == 0 && len(dep.FromField) > 0 || len(dep.ToField) > 0 && len(dep.FromField) == 0 {
				return fmt.Errorf("Resource %s incorrect specification of dependency on %s, fix FromField, ToField", r.Name(), dep.FromResource)
			}
			if err := checkField(r, dep.ToField); err != nil {
				return err
			}
			if err := checkField(cache[dep.FromResource], dep.FromField); err != nil {
				return err
			}
		}
	}

	return nil
}

func checkField(r Resource, field string) error {
	if len(field) == 0 {
		return nil
	}

	if reflect.ValueOf(r).Elem().Type().Name() == "protoBuilder" {
		if !reflect.ValueOf(r).Elem().FieldByName("UDef").Elem().Elem().FieldByName(field).IsValid() {
			return fmt.Errorf("in %s embedded Resource did not find field %s", r.Name(), field)
		}
	} else {
		if !reflect.ValueOf(r).Elem().FieldByName(field).IsValid() {
			return fmt.Errorf("in %s Resource did not find field %s", r.Name(), field)
		}
	}
	return nil
}

func copyValue(to Resource, toField string, from Resource, fromField string) {
	if len(toField) == 0 || len(fromField) == 0 {
		return
	}

	var fromValue reflect.Value
	if reflect.ValueOf(from).Elem().Type().Name() == "protoBuilder" {
		fromValue = reflect.ValueOf(from).Elem().FieldByName("UDef").Elem().Elem().FieldByName(fromField)
	} else {
		fromValue = reflect.ValueOf(from).Elem().FieldByName(fromField)
	}
	if reflect.ValueOf(to).Elem().Type().Name() == "protoBuilder" {
		reflect.ValueOf(to).Elem().FieldByName("UDef").Elem().Elem().FieldByName(toField).Set(fromValue)
	} else {
		reflect.ValueOf(to).Elem().FieldByName(toField).Set(fromValue)
	}
}

func createSync(resources []Resource, g *graph) (map[string]string, error) {
	ordered := sort(g)

	var err error

	resourcesLeft := len(ordered)
	maxAttempts := len(ordered)

	buildCache := map[string]Resource{}
	status := map[string]string{}
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
				if _, found := buildCache[dep.FromResource]; !found {
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
				c <- execute(b, buildCache)
			}(resources[i], output[i])
		}

		wg.Wait()

		errs := ErrorMap{}
		for i, c := range output {
			e := <-c

			if len(e.status) > 0 {
				status[resources[i].Name()] = e.status
			}

			if e.result != nil {
				logger("error executing resource", "resource", resources[i], "error", e)
				errs[i] = e.result
				continue
			}

			name := resources[i].Name()
			buildCache[name] = resources[i]
		}

		resourcesLeft -= len(execList) - errs.Size()
		err = errs.Get()
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return status, err
}

func execute(r Resource, cache map[string]Resource) builderOutput {
	for _, dep := range r.Dependencies() {
		copyValue(r, dep.ToField, cache[dep.FromResource], dep.FromField)
	}

	out, err := r.Update()
	return builderOutput{out, err}
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
