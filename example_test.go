package graph

import (
	"context"
	"fmt"
)

type MyFactory struct {
	ctxt context.Context
}

type Kinesis struct {
	*Resource
	ctxt context.Context
	bag  interface{}
}

func (k *Kinesis) Get() *Resource                           { return k.Resource }
func (k *Kinesis) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *Kinesis) Delete() error                            { return nil }

type Dynamo struct {
	*Resource
	ctxt context.Context
}

func (k *Dynamo) Get() *Resource                           { return k.Resource }
func (k *Dynamo) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *Dynamo) Delete() error                            { return nil }

type Deployment struct {
	*Resource
	ctxt context.Context
}

func (k *Deployment) Get() *Resource                           { return k.Resource }
func (k *Deployment) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *Deployment) Delete() error                            { return nil }

func (f *MyFactory) Create(r *Resource) Builder {
	switch r.Type {
	case "kinesis":
		return &Kinesis{r, f.ctxt, nil}
	case "dynamo":
		return &Dynamo{r, f.ctxt}
	case "deployment":
		return &Deployment{r, f.ctxt}
	}
	return nil
}

func Example_basic() {
	factory := &MyFactory{context.Background()}

	mykin := "mykin"

	resources := []*Resource{{
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

	err := Sync(resources, false, factory)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}
}
