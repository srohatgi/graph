package graph_test

import (
	"context"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/srohatgi/graph"
)

const spec = `
kinesis:
- name: mykin
  streamname: myEventStream
deployment:
- name: mydep
  dependencies:
  - fromresource: mykin
    fromfield: Arn
    tofield: KinesisArn
dynamo:
- name: mydyn
  tablename: myDynamoTable
`

const debugGraphLib = false

type factory struct {
	Kinesis    []*Kinesis
	Deployment []*Deployment
	Dynamo     []*Dynamo
}

func new(data string) (*factory, error) {
	var err error
	f := factory{}
	err = yaml.Unmarshal([]byte(data), &f)
	if err != nil {
		return nil, err
	}

	return &f, nil
}

func (f *factory) build() []graph.Resource {
	resources := []graph.Resource{}
	for _, k := range f.Kinesis {
		resources = append(resources, k)
	}
	for _, d := range f.Dynamo {
		resources = append(resources, d)
	}
	for _, d := range f.Deployment {
		resources = append(resources, d)
	}

	return resources
}

/*
This example shows basic resource synchronization. There are three
different resources that we need to build: an AWS Kinesis stream, an
Aws Dynamo DB table, and finally a Kubernetes deployment of a micro-
service that depends on both of the other resources being created
properly.
*/
func Example_usage() {

	f, err := new(spec)
	if err != nil {
		fmt.Printf("error creating factory: %v\n", err)
	}

	resources := f.build()

	if debugGraphLib {
		myprint := func(in ...interface{}) {
			fmt.Println(in...)
		}

		graph.WithLogger(myprint)
	}

	//fmt.Printf("factory: %v\n", f)

	ctxt := context.WithValue(context.Background(), graph.SyncBag, map[string]string{"namespace": "myns"})

	status, err := graph.Sync(ctxt, resources, false)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}

	fmt.Printf("deployment status = %s\n", status["mydep"])
	// Output:
	// deployment status = successfully reading hello123 in myns
}

// AWS Kinesis resource definition
type Kinesis struct {
	graph.Depends `yaml:",inline"`
	StreamName    string
	Arn           string
}

func (kin *Kinesis) Update(ctxt context.Context) (string, error) {
	kin.Arn = "hello123"
	return "", nil
}
func (kin *Kinesis) Delete(ctxt context.Context) error {
	return nil
}

// AWS Dynamo DB resource definition
type Dynamo struct {
	graph.Depends `yaml:",inline"`
	TableName     string
}

func (dyn *Dynamo) Update(ctxt context.Context) (string, error) {
	return "", nil
}
func (dyn *Dynamo) Delete(ctxt context.Context) error {
	return nil
}

// Kubernetes Deployment resource definition
type Deployment struct {
	graph.Depends `yaml:",inline"`
	KinesisArn    string
}

func (dep *Deployment) Update(ctxt context.Context) (string, error) {
	crd, ok := ctxt.Value(graph.SyncBag).(map[string]string)
	if !ok {
		return "", fmt.Errorf("unable to get crd info")
	}
	// use KinesisArn
	return "successfully reading " + dep.KinesisArn + " in " + crd["namespace"], nil
}
func (dep *Deployment) Delete(ctxt context.Context) error {
	return nil
}
