package graph

import (
	"context"
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

func TestCheckField(t *testing.T) {
	ctxt := context.Background()
	kinesisResource := MakeResource("mykin", nil, &kinesis{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })

	if checkField(kinesisResource, "Arn") != nil {
		t.Fatal("Arn field exists in kinesisResource")
	}
	if checkField(kinesisResource, "Bad") == nil {
		t.Fatal("Bad field does not exist in kinesisResource")
	}
}

func TestCopyValue(t *testing.T) {
	ctxt := context.Background()

	arn := "hello123"

	kinesisResource := MakeResource("mykin", nil, &kinesis{ctxt, arn}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", []Dependency{{"mykin", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { d := x.(*deployment); return d.KinesisArn, nil }, func(x interface{}) error { return nil })

	copyValue(deploymentResource, "KinesisArn", kinesisResource, "Arn")

	out, err := deploymentResource.Update(context.Background())

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

	arn := "hello123"

	kinesisResource := MakeResource(mykin, nil, &kinesis{ctxt, arn}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	dynamoResource := MakeResource("mydyn", nil, &dynamo{ctxt: ctxt}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", []Dependency{{"mykin", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { d := x.(*deployment); return d.KinesisArn, nil }, func(x interface{}) error { return nil })

	resources := []Resource{kinesisResource, dynamoResource, deploymentResource}

	lib := New(&Opts{CustomLogger: t.Log})

	status, err := lib.Sync(ctxt, resources, false)

	if err != nil {
		t.Fatalf("unable to sync %v", err)
	}

	if status["mydep1"] != arn {
		t.Fatal("expected mydep1 status to return kinesis arn")
	}
}

func TestInvalidDependency(t *testing.T) {

	mykin := "mykin"

	ctxt := context.Background()

	arn := "hello123"

	kinesisResource := MakeResource(mykin, nil, &kinesis{ctxt, arn}, func(x interface{}) (string, error) { return "", nil }, func(x interface{}) error { return nil })
	deploymentResource := MakeResource("mydep1", []Dependency{{"mykin2", "Arn", "KinesisArn"}}, &deployment{ctxt: ctxt}, func(x interface{}) (string, error) { d := x.(*deployment); return d.KinesisArn, nil }, func(x interface{}) error { return nil })

	resources := []Resource{kinesisResource, deploymentResource}

	lib := New(&Opts{CustomLogger: t.Log})

	_, err := lib.Sync(ctxt, resources, false)

	if err == nil {
		t.Fatalf("expected sync to fail with invalid dependency")
	}
}
