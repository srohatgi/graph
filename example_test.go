package graph_test

import (
	"context"
	"fmt"

	yaml "gopkg.in/yaml.v2"

	"github.com/imdario/mergo"
	"github.com/srohatgi/graph"
)

const spec = `
kinesis:
- name: mykin
  shardcount: 5
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

const overrideSpec = `
kinesis:
- name: mykin
  shardcount: 10
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

func (f *factory) applyOverrides(defaults *factory) {
	for _, dest := range f.Kinesis {
		for _, src := range defaults.Kinesis {
			if dest.ResourceName() == src.ResourceName() {
				mergo.Merge(dest, *src, mergo.WithOverride)
			}
		}
	}
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

	envFactory, err := new(overrideSpec)
	if err != nil {
		fmt.Printf("error creating override factory: %v\n", err)
	}

	f.applyOverrides(envFactory)

	resources := f.build()

	myprint := func(in ...interface{}) {
		if debugGraphLib {
			fmt.Println(in...)
		}
	}

	lib := graph.New(&graph.Opts{CustomLogger: myprint})

	//fmt.Printf("factory: %v\n", f)

	ctxt := context.WithValue(context.Background(), graph.SyncBag, map[string]string{"namespace": "myns"})

	status, err := lib.Sync(ctxt, resources, false)
	if err != nil {
		if em, ok := err.(graph.ErrorMapper); ok {
			for resourceName, err := range em.ErrorMap() {
				fmt.Printf("resource %s creation had error %v\n", resourceName, err)
			}
		} else {
			fmt.Printf("unable to sync resources, error = %v\n", err)
		}
	}

	fmt.Printf("kinesis status = %s\n", status["mykin"])
	fmt.Printf("deployment status = %s\n", status["mydep"])
	// Output:
	// kinesis status = successfully created stream myEventStream with 10 shards
	// deployment status = successfully reading from stream arn hello123 in myns
}

// AWS Kinesis resource definition
type Kinesis struct {
	graph.Depends `yaml:",inline"`
	ShardCount    int
	StreamName    string
	Arn           string
}

func (kin *Kinesis) Update(ctxt context.Context) (string, error) {
	kin.Arn = "hello123"
	return fmt.Sprintf("successfully created stream %s with %d shards", kin.StreamName, kin.ShardCount), nil
}
func (kin *Kinesis) Delete(ctxt context.Context) error {
	return nil
}

func (kin *Kinesis) IsReady(ctxt context.Context) bool {
	return true
}

func (kin *Kinesis) Get(ctxt context.Context) (interface{}, error) {
	return nil, nil
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

func (dyn *Dynamo) IsReady(ctxt context.Context) bool {
	return true
}

func (dyn *Dynamo) Get(ctxt context.Context) (interface{}, error) {
	return nil, nil
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
	return fmt.Sprintf("successfully reading from stream arn %s in %s", dep.KinesisArn, crd["namespace"]), nil
}
func (dep *Deployment) Delete(ctxt context.Context) error {
	return nil
}

func (dep *Deployment) IsReady(ctxt context.Context) bool {
	return true
}

func (dep *Deployment) Get(ctxt context.Context) (interface{}, error) {
	return nil, nil
}
