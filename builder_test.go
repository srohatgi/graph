package graph

import (
	"context"
	"fmt"
	"testing"
)

type kinesis struct {
	ctxt context.Context
	Arn  interface{}
}

type dynamo struct {
	ctxt context.Context
}

type deployment struct {
	ctxt       context.Context
	KinesisArn interface{}
}

func TestSync(t *testing.T) {
	mykin := "mykin"

	ctxt := context.Background()

	kinesisResource := MakeResource(mykin, "kinesis", nil, &kinesis{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	dynamoResource := MakeResource("mydyn", "dynamo", nil, &dynamo{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", "deployment", []Dependency{{"mykin", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })

	resources := []Resource{kinesisResource, dynamoResource, deploymentResource}

	WithLogger(t.Log)

	err := Sync(resources, false)

	if err != nil {
		fmt.Print(err)
		t.Fatalf("unable to sync %v", err)
	}

}
