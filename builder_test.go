package graph

import "testing"

type factory struct{}

type kinesis struct {
	*Resource
}

func (k *kinesis) Get() *Resource                           { return k.Resource }
func (k *kinesis) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *kinesis) Delete() error                            { return nil }

type dynamo struct {
	*Resource
}

func (k *dynamo) Get() *Resource                           { return k.Resource }
func (k *dynamo) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *dynamo) Delete() error                            { return nil }

type deployment struct {
	*Resource
}

func (k *deployment) Get() *Resource                           { return k.Resource }
func (k *deployment) Update(in []Property) ([]Property, error) { return nil, nil }
func (k *deployment) Delete() error                            { return nil }

func (f *factory) Create(r *Resource) Builder {
	switch r.Type {
	case "kinesis":
		return &kinesis{r}
	case "dynamo":
		return &dynamo{r}
	case "deployment":
		return &deployment{r}
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
		DependsOn: []Dependency{{mykin, []Property{{"ARN", ""}}}},
	}}

	f := &factory{}

	err := Sync(resources, false, f)

	if err != nil {
		t.Fatalf("unable to sync %v", err)
	}

}
