// Package graph may be useful for developers dealing with creating, updating and tearing down resources.
// The package assumes that a declarative model of resources exists.
//
// Developers start by mapping thier resources to the provided Resource{} data structure. Developers
// also need to define a Factory{} interface for creating resource Builder{} interface.
//
// All this mapping leads to calling a single function which handles creation and deletion of resources:
//  err := Sync(resources, false, factory)
// 
// Sync method executes resource creation in parallel if it can. ErrorSlice is a new type that developers 
// may use to get fine grained information on the issues.
//  if es, ok := err.(*ErrorSlice); ok {
//    for resourceIndex, err := range *es {
//      fmt.Printf("resource %d creation had error %v\n", resourceIndex, err)
//    }
//  }
package graph

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// graph data type
type graph struct {
	v   int
	adj [][]int
}

// new graph with v vertices
func newGraph(v int) *graph {
	return &graph{v: v, adj: make([][]int, v)}
}

// NewFromReader assumes number of vertices, number of edges, and then each edge per line
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
