package graph_test

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/srohatgi/graph"
)

const data = `
kinesis:
- resourcename: mykin
  streamname: myEventStream
deployment:
- resourcename: mydep
  resourcedependencies:
  - fromresource: mykin
    fromfield: Arn
    tofield: KinesisArn
dynamo:
- resourcename: mydyn
  tablename: myDynamoTable
`

const debugGraphLib = false

type CRD struct {
	Kinesis    []Kinesis
	Deployment []Deployment
	Dynamo     []Dynamo
}

/*
This example shows basic resource synchronization. There are three
different resources that we need to build: an AWS Kinesis stream, an
Aws Dynamo DB table, and finally a Kubernetes deployment of a micro-
service that depends on both of the other resources being created
properly.
*/
func Example_usage() {

	crd := CRD{}

	err := yaml.Unmarshal([]byte(data), &crd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
	}

	resources := []graph.Resource{}
	for _, k := range crd.Kinesis {
		resources = append(resources, &k)
	}
	for _, d := range crd.Deployment {
		resources = append(resources, &d)
	}
	for _, d := range crd.Dynamo {
		resources = append(resources, &d)
	}

	myprint := func(in ...interface{}) {
		fmt.Println(in...)
	}

	if debugGraphLib {
		graph.WithLogger(myprint)
	}

	status, err := graph.Sync(resources, false)
	if err != nil {
		fmt.Printf("unable to sync resources, error = %v\n", err)
	}

	fmt.Printf("deployment status = %s\n", status["mydep"])
	// Output:
	// deployment status = hello123
}

// AWS Kinesis resource definition
type Kinesis struct {
	ResourceName         string
	ResourceDependencies []graph.Dependency
	StreamName           string
	Arn                  string
}

func (kin *Kinesis) Name() string {
	return kin.ResourceName
}
func (kin *Kinesis) Dependencies() []graph.Dependency {
	return kin.ResourceDependencies
}
func (kin *Kinesis) Update() (string, error) {
	kin.Arn = "hello123"
	return "", nil
}
func (kin *Kinesis) Delete() error {
	return nil
}

// AWS Dynamo DB resource definition
type Dynamo struct {
	ResourceName         string
	ResourceDependencies []graph.Dependency
	TableName            string
}

func (dyn *Dynamo) Name() string {
	return dyn.ResourceName
}
func (dyn *Dynamo) Dependencies() []graph.Dependency {
	return dyn.ResourceDependencies
}
func (dyn *Dynamo) Update() (string, error) {
	return "", nil
}
func (dyn *Dynamo) Delete() error {
	return nil
}

// Kubernetes Deployment resource definition
type Deployment struct {
	ResourceName         string
	ResourceDependencies []graph.Dependency
	KinesisArn           string
}

func (dep *Deployment) Name() string {
	return dep.ResourceName
}
func (dep *Deployment) Dependencies() []graph.Dependency {
	return dep.ResourceDependencies
}
func (dep *Deployment) Update() (string, error) {
	// use KinesisArn
	return dep.KinesisArn, nil
}
func (dep *Deployment) Delete() error {
	return nil
}
