// Package graph may be useful for developers tasked with storage, compute and
// network management for cloud microservices.  The library provides types and
// functions that enable a modular, extensible programming model.
//
// Types and Values
//
// Resource interface is a declarative abstraction of a storage, compute or network
// service. Resources may have a Dependency order of creation and deletion.
// Idiomatic resources may have a single backing structure for fulfilling the
// interface. The library utilizes this backing structure to locate and inject
// public properties of a resource during execution.
//
// Functions
//
// The library manages a collection of related resources at a given time.
// The Sync() function provides a method for managing a collection of resources:
//
//  status, err := lib.Sync(context, resources, false) // refer to signature below
//
// The library tries to execute multiple resources concurrently. There is a handy
// ErrorMapper interface that allows developers to query resource specific errors.
//
// Use the following code snippet:
//  if em, ok := err.(ErrorMapper); ok {
//    for resourceName, err := range em.ErrorMap() {
//      fmt.Printf("resource %s creation had error %v\n", resourceName, err)
//    }
//  }
//
package graph

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Opts captures customizable functionality like logging
type Opts struct {
	CustomLogger func(args ...interface{})
}

// New creates an instance object
func New(opts *Opts) *Lib {
	lib := &Lib{
		logger: func(args ...interface{}) {},
	}
	if opts != nil && opts.CustomLogger != nil {
		lib.logger = opts.CustomLogger
	}

	return lib
}

// Lib object is required for using the library
type Lib struct {
	logger func(args ...interface{})
}

// graph data type
type graph struct {
	v   int
	adj [][]int
}

// new graph with v vertices
func newGraph(v int) *graph {
	return &graph{v: v, adj: make([][]int, v)}
}

// newFromReader assumes number of vertices, number of edges, and then each edge per line
func newFromReader(r io.Reader) (*graph, error) {
	scanner := bufio.NewScanner(r)

	var g *graph
	var err error
	edges := -1
	for scanner.Scan() {
		if err != nil {
			break
		}

		s := scanner.Text()

		if g == nil {
			var v int
			v, err = strconv.Atoi(s)
			if err == nil {
				g = newGraph(v)
			}
		} else if edges == -1 {
			edges, err = strconv.Atoi(s)
		} else if edges > 0 {
			var v1, w1, nums int
			nums, err = fmt.Sscanf(s, "%d %d", &v1, &w1)
			if nums != 2 {
				err = errors.New("illegal edge: " + s)
			} else if err == nil {
				g.addEdge(v1, w1)
			}
			edges--
		}
	}

	if scanner.Err() != nil && err == nil {
		err = scanner.Err()
	}

	return g, err
}

// Vertices in the graph
func (g *graph) vertices() int {
	return g.v
}

// adjascent vertices to v1
func (g *graph) adjascent(v1 int) []int {
	return g.adj[v1]
}

// AddEdge (v1, w1)
func (g *graph) addEdge(v1, w1 int) {
	g.adj[v1] = append(g.adj[v1], w1)
}

// String representation
func (g *graph) String() string {
	return fmt.Sprintf("v=%v, adj=%v", g.v, g.adj)
}
