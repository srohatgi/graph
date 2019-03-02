package graph

import (
	"context"
	"fmt"
	"testing"
)

type kinesis struct {
	ctxt context.Context
	Arn  string
}

type dynamo struct {
	ctxt context.Context
}

type deployment struct {
	ctxt       context.Context
	KinesisArn string
}

func TestCopyValue(t *testing.T) {
	WithLogger(t.Log)
	ctxt := context.Background()

	arn := "hello123"

	kinesisResource := MakeResource("mykin", "kinesis", nil, &kinesis{ctxt, arn}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", "deployment", []Dependency{{"mykin", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { d := x.(*deployment); return d.KinesisArn, nil }, func(x interface{}) error { return nil })

	copyValue(deploymentResource, "KinesisArn", kinesisResource, "Arn")

	out, err := deploymentResource.Update()

	if err != nil {
		t.Fatalf("error calling Update! err = %v", err)
	}

	if out != arn {
		t.Fatalf("expected arn to match!")
	}

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
