package graph

import (
	"context"
	"testing"
)

const (
	arnProperty = "ARN"
	streamName  = "hello123"
)

type factory struct {
	ctxt context.Context
}

type kinesis struct {
	*Resource
	ctxt context.Context
	bag  interface{}
}

func (k *kinesis) Get() *Resource { return k.Resource }
func (k *kinesis) Update(in []Property) ([]Property, error) {
	return []Property{{arnProperty, streamName}}, nil
}
func (k *kinesis) Delete() error { return nil }

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

	status, err := Sync(resources, false, f)

	if err != nil {
		t.Fatalf("unable to sync %v", err)
	}

	if status[mykin][0].Value != streamName {
		t.Fatal("kinesis stream name not sent out")
	}

}
