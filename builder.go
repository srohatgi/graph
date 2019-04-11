package graph

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type bag string

// SyncBag allows for retrieving global context values from a context
const SyncBag bag = "crd"

// Depends is a convenience structure used for capturing resource dependencies.
type Depends struct {
	Name         string
	Dependencies []Dependency
}

// Dependency specifies a single dependency
type Dependency struct {
	// FromResource is another resource specified in the same slice.
	FromResource string
	// FromField is a public field from the struct implementing the Resource.
	FromField string
	// ToField is a public field in the current Resource's implementing struct.
	ToField string
}

// ResourceName convenience function
func (dep *Depends) ResourceName() string {
	return dep.Name
}

// ResourceDependencies convenience function
func (dep *Depends) ResourceDependencies() []Dependency {
	return dep.Dependencies
}

// Resource is an abstract declarative definition for compute, storage and network services.
// Examples: AWS Kinesis, AWS CloudFormation, Kubernetes Deployment etc.
type Resource interface {
	Depender
	Builder
}

type builderOutput struct {
	status string
	result error
}

// Depender captures dependencies between resources
type Depender interface {
	// Get retrieves underlying Resource instance name. This allows creation
	// of multiple resources of the same Type.
	ResourceName() string
	// Dependencies fetches a given Resource's dependency list.
	ResourceDependencies() []Dependency
}

// Builder allows resources to be created/ deleted
type Builder interface {
	// Delete the Resource.
	Delete(ctxt context.Context) error
	// Update or if not existing, create the Resource.
	Update(ctxt context.Context) (string, error)
}

type protoBuilder struct {
	Name         string
	Dependencies []Dependency
	UDef         interface{}
	UpdFn        func(interface{}) (string, error)
	DelFn        func(interface{}) error
}

func (p *protoBuilder) ResourceName() string                        { return p.Name }
func (p *protoBuilder) Update(ctxt context.Context) (string, error) { return p.UpdFn(p.UDef) }
func (p *protoBuilder) Delete(ctxt context.Context) error           { return p.DelFn(p.DelFn) }
func (p *protoBuilder) ResourceDependencies() []Dependency          { return p.Dependencies }

// MakeResource is a convenient utility to create Resource's in a cheap way.
// NOTE: uDef is a custom generic struct that is injected into updFn & delFn
func MakeResource(name string, dependencies []Dependency, uDef interface{}, updFn func(interface{}) (string, error), delFn func(interface{}) error) Resource {
	return &protoBuilder{name, dependencies, uDef, updFn, delFn}
}

// Sync method uses the Resource slice to generate a DAG. The DAG is processed based on the value
// of toDelete flag. Resources may be processed concurrently. Processed resources may return a status
// string and or an error. The function collects these and aggregates them in respective maps keyed by
// resource names.
func (lib *Lib) Sync(ctxt context.Context, resources []Resource, toDelete bool) (map[string]string, error) {
	err := check(resources)
	if err != nil {
		return nil, err
	}

	g := buildGraph(resources)

	lib.logger("starting sync")

	if toDelete {
		return nil, lib.deleteSync(ctxt, resources, g)
	}

	return lib.createSync(ctxt, resources, g)
}

// check that resources have correct dependencies
func check(resources []Resource) error {
	cache := map[string]Resource{}

	for _, r := range resources {
		cache[r.ResourceName()] = r
	}

	for n, r := range cache {
		if reflect.ValueOf(r).Kind() != reflect.Ptr {
			return fmt.Errorf("expected %s Resource to be implemented with a pointer to struct", n)
		}

		if reflect.ValueOf(r).Elem().Kind() != reflect.Struct {
			return fmt.Errorf("expected %s Resource to be implemented using a pointer to struct", n)
		}

		// validate each dependency
		for _, dep := range r.ResourceDependencies() {
			if len(dep.ToField) == 0 && len(dep.FromField) > 0 || len(dep.ToField) > 0 && len(dep.FromField) == 0 {
				return fmt.Errorf("Resource %s incorrect specification of dependency on %s, fix FromField, ToField", r.ResourceName(), dep.FromResource)
			}
			if err := checkField(r, dep.ToField); err != nil {
				return err
			}
			if _, ok := cache[dep.FromResource]; !ok {
				return fmt.Errorf("Dependent resource %s doesn't exist", dep.FromResource)
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
			return fmt.Errorf("in %s embedded Resource did not find field %s", r.ResourceName(), field)
		}
	} else {
		if !reflect.ValueOf(r).Elem().FieldByName(field).IsValid() {
			return fmt.Errorf("in %s Resource did not find field %s", r.ResourceName(), field)
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

func (lib *Lib) createSync(ctxt context.Context, resources []Resource, g *graph) (map[string]string, error) {
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
			if _, alreadyExecuted := buildCache[res.ResourceName()]; alreadyExecuted {
				lib.logger("already executed", i)
				continue
			}

			ready := true
			for _, dep := range res.ResourceDependencies() {
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

		lib.logger("executing ", execList)
		for _, i := range execList {
			wg.Add(1)
			output[i] = make(chan builderOutput, 1)

			go func(b Resource, c chan builderOutput) {
				defer wg.Done()
				c <- execute(ctxt, b, buildCache)
			}(resources[i], output[i])
		}

		wg.Wait()

		errs := errorMap{}
		for i, c := range output {
			e := <-c

			if len(e.status) > 0 {
				status[resources[i].ResourceName()] = e.status
			}

			if e.result != nil {
				lib.logger("error executing resource", "resource", resources[i], "error", e)
				errs[resources[i].ResourceName()] = e.result
				continue
			}

			name := resources[i].ResourceName()
			buildCache[name] = resources[i]
		}

		resourcesLeft -= len(execList) - len(errs)

		if len(errs) > 0 {
			err = errs
		}
	}

	if resourcesLeft > 0 && err == nil {
		err = errors.New("max attempts at computing resources exhausted, giving up")
	}

	return status, err
}

func execute(ctxt context.Context, r Resource, cache map[string]Resource) builderOutput {
	for _, dep := range r.ResourceDependencies() {
		copyValue(r, dep.ToField, cache[dep.FromResource], dep.FromField)
	}

	out, err := r.Update(ctxt)
	return builderOutput{out, err}
}

func reverse(in []int) {
	for i := len(in)/2 - 1; i >= 0; i-- {
		opp := len(in) - 1 - i
		in[i], in[opp] = in[opp], in[i]
	}
}

func (lib *Lib) deleteSync(ctxt context.Context, resources []Resource, g *graph) error {
	order := sort(g)
	reverse(order)

	lib.logger("order of deletion", order)

	var err error

	for _, i := range order {
		err = resources[i].Delete(ctxt)
		if err != nil {
			break
		}
	}

	return err
}
