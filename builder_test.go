package graph

import "testing"

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

	err := Sync(resources, false)

	if err != nil {
		t.Fatalf("unable to sync %v", err)
	}

}
