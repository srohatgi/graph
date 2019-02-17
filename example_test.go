package graph_test

import (
	"context"
	"fmt"

	"github.com/srohatgi/graph"
)

// MyFactory keeps a context object
type MyFactory struct {
	ctxt context.Context
}

func (f *MyFactory) Create(r *graph.Resource) graph.Builder {
	switch r.Type {
	case "kinesis":
		// a Builder may be injected with any user defined types
		// here we are passing nil
		return &Kinesis{r, f.ctxt, nil}
	case "dynamo":
		return &Dynamo{r, f.ctxt}
	case "deployment":
		return &Deployment{r, f.ctxt}
	}
	return nil
}

/*
This example shows a basic resource synchronization. There are three
different resources that we need to build: an AWS Kinesis stream, an
Aws Dynamo DB table, and finally a Kubernetes deployment of a micro-
service that depends on both of the other resources being created
properly.
*/
func Example_basic() {
	factory := &MyFactory{context.Background()}

	mykin := "mykin"

	resources := []*graph.Resource{{
		Name: mykin,
		Type: "kinesis",
	}, {
		Name: "mydyn",
		Type: "dynamo",
	}, {
		Name:      "mydep1",
		Type:      "deployment",
		DependsOn: []string{mykin},
	}}

	err := graph.Sync(resources, false, factory)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}
}

type Kinesis struct {
	*graph.Resource
	ctxt context.Context
	bag  interface{}
}

func (k *Kinesis) Get() *graph.Resource {
	return k.Resource
}
func (k *Kinesis) Update(in []graph.Property) ([]graph.Property, error) {
	return nil, nil
}
func (k *Kinesis) Delete() error {
	return nil
}

type Dynamo struct {
	*graph.Resource
	ctxt context.Context
}

func (k *Dynamo) Get() *graph.Resource {
	return k.Resource
}
func (k *Dynamo) Update(in []graph.Property) ([]graph.Property, error) {
	return nil, nil
}
func (k *Dynamo) Delete() error {
	return nil
}

type Deployment struct {
	*graph.Resource
	ctxt context.Context
}

func (k *Deployment) Get() *graph.Resource {
	return k.Resource
}
func (k *Deployment) Update(in []graph.Property) ([]graph.Property, error) {
	return nil, nil
}
func (k *Deployment) Delete() error {
	return nil
}
