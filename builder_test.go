package graph

import "testing"

type factory struct{}

type kinesis struct {
	*Resource
}

func (k *kinesis) Update(cache BuildCache) error { return nil }
func (k *kinesis) Delete(cache BuildCache) error { return nil }

type dynamo struct {
	*Resource
}

func (k *dynamo) Update(cache BuildCache) error { return nil }
func (k *dynamo) Delete(cache BuildCache) error { return nil }

type deployment struct {
	*Resource
}

func (k *deployment) Update(cache BuildCache) error { return nil }
func (k *deployment) Delete(cache BuildCache) error { return nil }

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
