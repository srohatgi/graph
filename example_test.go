package graph_test

import (
	"context"
	"fmt"

	"github.com/srohatgi/graph"
)

// MyFactory keeps a context object
type MyFactory struct {
	ctxt          context.Context
	kinesisCustom *myUserDefinedType
}

// Create satisfies the graph.Factory interface
func (f *MyFactory) Create(r *graph.Resource) graph.Builder {
	switch r.Type {
	case "kinesis":
		// a Builder may be injected with any user defined types
		// here we are passing a custom myUserDefinedType struct
		updFn := func(u interface{}, in []graph.Property) ([]graph.Property, error) {
			// use the u.streamName to construct your kinesis stream
			return nil, nil
		}
		delFn := func(u interface{}) error {
			// use the u.streamName to delete the stream
			return nil
		}
		return graph.MakeBuilder(r, f.kinesisCustom, updFn, delFn)
	case "dynamo":
		return &Dynamo{r, f.ctxt}
	case "deployment":
		return &Deployment{r, f.ctxt}
	}
	return nil
}

/*
This example shows basic resource synchronization. There are three
different resources that we need to build: an AWS Kinesis stream, an
Aws Dynamo DB table, and finally a Kubernetes deployment of a micro-
service that depends on both of the other resources being created
properly.
*/
func Example_usage() {
	ctxt := context.Background()
	factory := &MyFactory{ctxt, &myUserDefinedType{ctxt, "myEventStream"}}

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

	_, err := graph.Sync(resources, false, factory)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}
}

// AWS Kinesis resource definition
type myUserDefinedType struct {
	ctxt       context.Context
	streamName string
}

// AWS Dynamo DB resource definition
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

// Kubernetes Deployment resource definition
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
