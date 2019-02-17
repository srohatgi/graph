package graph

import (
	"context"
	"fmt"
	"testing"
)

type factory struct {
	ctxt context.Context
}

type kinesis struct {
	*Resource
	ctxt context.Context
	bag  interface{}
}

func (k *kinesis) Get() *Resource                           { return k.Resource }
func (k *kinesis) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *kinesis) Delete() error                            { return nil }

type dynamo struct {
	*Resource
	ctxt context.Context
}

func (k *dynamo) Get() *Resource                           { return k.Resource }
func (k *dynamo) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *dynamo) Delete() error                            { return nil }

type deployment struct {
	*Resource
	ctxt context.Context
}

func (k *deployment) Get() *Resource                           { return k.Resource }
func (k *deployment) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *deployment) Delete() error                            { return nil }

func (f *factory) Example(r *Resource) Builder {
	switch r.Type {
	case "kinesis":
		// inject context, arbitrary parameters
		return &kinesis{r, f.ctxt, nil}
	case "dynamo":
		return &dynamo{r, f.ctxt}
	case "deployment":
		return &deployment{r, f.ctxt}
	}
	return nil
}

func (f *factory) Create(r *Resource) Builder {
	switch r.Type {
	case "kinesis":
		return &kinesis{r, f.ctxt, nil}
	case "dynamo":
		return &dynamo{r, f.ctxt}
	case "deployment":
		return &deployment{r, f.ctxt}
	}
	return nil
}

func TestSync(t *testing.T) {
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

	f := &factory{}

	WithLogger(t.Log)

	err := Sync(resources, false, f)

	if err != nil {
		fmt.Print(err)
		t.Fatalf("unable to sync %v", err)
	}

}
