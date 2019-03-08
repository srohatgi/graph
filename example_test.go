package graph_test

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/srohatgi/graph"
)

const data = `
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
	Name         string
	Dependencies []graph.Dependency
	StreamName   string
	Arn          string
}

func (kin *Kinesis) ResourceName() string {
	return kin.Name
}
func (kin *Kinesis) ResourceDependencies() []graph.Dependency {
	return kin.Dependencies
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
	Name         string
	Dependencies []graph.Dependency
	TableName    string
}

func (dyn *Dynamo) ResourceName() string {
	return dyn.Name
}
func (dyn *Dynamo) ResourceDependencies() []graph.Dependency {
	return dyn.Dependencies
}
func (dyn *Dynamo) Update() (string, error) {
	return "", nil
}
func (dyn *Dynamo) Delete() error {
	return nil
}

// Kubernetes Deployment resource definition
type Deployment struct {
	Name         string
	Dependencies []graph.Dependency
	KinesisArn   string
}

func (dep *Deployment) ResourceName() string {
	return dep.Name
}
func (dep *Deployment) ResourceDependencies() []graph.Dependency {
	return dep.Dependencies
}
func (dep *Deployment) Update() (string, error) {
	// use KinesisArn
	return dep.KinesisArn, nil
}
func (dep *Deployment) Delete() error {
	return nil
}
