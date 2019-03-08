package graph

import (
	"context"
	"testing"
)

func TestBuildGraph(t *testing.T) {

	ctxt := context.Background()
	mykin := "mykin"

	kinesisResource := MakeResource(mykin, nil, &kinesis{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	dynamoResource := MakeResource("mydyn", nil, &dynamo{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", []Dependency{{"mykin", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })

	resources := []Resource{kinesisResource, dynamoResource, deploymentResource}

	g := buildGraph(resources)

	t.Logf("graph = %v\n", g)

	if g.vertices() != 3 {
		t.Fatal("vertices incorrect")
	}

	if len(g.adjascent(0)) != 1 || g.adjascent(0)[0] != 2 {
		t.Fatal("incorrect edges")
	}
}
