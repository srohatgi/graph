package graph

import "testing"

type factory struct{}

// Kinesis is aws data stream
type Kinesis struct {
	*Resource
}

func (k *Kinesis) Update(cache BuildCache) error { return nil }
func (k *Kinesis) Delete(cache BuildCache) error { return nil }

// Dynamo is aws data table
type Dynamo struct {
	*Resource
}

func (k *Dynamo) Update(cache BuildCache) error { return nil }
func (k *Dynamo) Delete(cache BuildCache) error { return nil }

// Deployment is kubernetes deployment
type Deployment struct {
	*Resource
}

func (k *Deployment) Update(cache BuildCache) error { return nil }
func (k *Deployment) Delete(cache BuildCache) error { return nil }

func (f *factory) Create(r *Resource) Builder {
	switch r.Type {
	case "kinesis":
		return &Kinesis{r}
	case "dynamo":
		return &Dynamo{r}
	case "deployment":
		return &Deployment{r}
	}
	return nil
}

func TestSync(t *testing.T) {

	resources := []*Resource{{
		Name: "mykin",
		Type: "kinesis",
	}, {
		Name: "mydyn",
		Type: "dynamo",
	}, {
		Name:      "mydep1",
		Type:      "deployment",
		DependsOn: []string{"mykin"},
	}}

	f := &factory{}

	err := Sync(resources, false, f)

	if err != nil {
		t.Fatalf("unable to sync %v", err)
	}

}
