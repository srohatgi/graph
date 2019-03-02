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
func (f *MyFactory) Create(resName, resType string, dependencies []graph.Dependency) graph.Resource {
	switch resType {
	case "kinesis":
		// a Builder may be injected with any user defined types
		// here we are passing a custom myUserDefinedType struct
		updFn := func(u interface{}) (string, error) {
			// use the u.streamName to construct your kinesis stream
			myks := u.(*myUserDefinedType)
			myks.Arn = "hello123"
			return "", nil
		}
		delFn := func(d interface{}) error {
			// use the d.streamName to delete the stream
			return nil
		}
		return graph.MakeResource(resName, resType, dependencies, f.kinesisCustom, updFn, delFn)
	case "dynamo":
		return &Dynamo{f.ctxt, resName, resType, dependencies}
	case "deployment":
		return &Deployment{ctxt: f.ctxt, resName: resName, resType: resType, dependencies: dependencies}
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
	factory := &MyFactory{ctxt, &myUserDefinedType{ctxt: ctxt, streamName: "myEventStream"}}

	mykin := "mykin"

	resources := []graph.Resource{
		factory.Create(mykin, "kinesis", nil),
		factory.Create("mydyn", "dynamo", nil),
		factory.Create("mydep1", "deployment", []graph.Dependency{{FromResource: mykin, FromField: "Arn", ToField: "KinesisArn"}}),
	}

	status, err := graph.Sync(resources, false)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}

	fmt.Printf("status = %v\n", status)
}

// AWS Kinesis resource definition
type myUserDefinedType struct {
	ctxt       context.Context
	streamName string
	Arn        string
}

// AWS Dynamo DB resource definition
type Dynamo struct {
	ctxt         context.Context
	resName      string
	resType      string
	dependencies []graph.Dependency
}

func (k *Dynamo) Name() string {
	return k.resName
}
func (k *Dynamo) Type() string {
	return k.resType
}
func (k *Dynamo) Dependencies() []graph.Dependency {
	return k.dependencies
}
func (k *Dynamo) Update() (string, error) {

	return "", nil
}
func (k *Dynamo) Delete() error {
	return nil
}

// Kubernetes Deployment resource definition
type Deployment struct {
	ctxt         context.Context
	resName      string
	resType      string
	dependencies []graph.Dependency
	KinesisArn   string
}

func (k *Deployment) Name() string {
	return k.resName
}
func (k *Deployment) Type() string {
	return k.resType
}
func (k *Deployment) Dependencies() []graph.Dependency {
	return k.dependencies
}
func (k *Deployment) Update() (string, error) {
	// use KinesisArn
	fmt.Printf("kinesis arn injected from kinesisBuilder is %s", k.KinesisArn)
	return "", nil
}
func (k *Deployment) Delete() error {
	return nil
}
