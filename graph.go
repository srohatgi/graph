// Package graph is a helper package for developers dealing with creating, updating and tearing down resources
// using kubernetes controllers.
package graph

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
)

// Graph data type
type Graph struct {
	v   int
	adj [][]int
}

// New Graph with v vertices
func New(v int) *Graph {
	return &Graph{v: v, adj: make([][]int, v)}
}

// NewFromReader assumes number of vertices, number of edges, and then each edge per line
func NewFromReader(r io.Reader) (*Graph, error) {
	scanner := bufio.NewScanner(r)

	var g *Graph
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
				g = New(v)
			}
		} else if edges == -1 {
			edges, err = strconv.Atoi(s)
		} else if edges > 0 {
			var v1, w1, nums int
			nums, err = fmt.Sscanf(s, "%d %d", &v1, &w1)
			if nums != 2 {
				err = errors.New("illegal edge: " + s)
			} else if err == nil {
				g.AddEdge(v1, w1)
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
func (g *Graph) Vertices() int {
	return g.v
}

// Adjascent vertices to v1
func (g *Graph) Adjascent(v1 int) []int {
	return g.adj[v1]
}

// AddEdge (v1, w1)
func (g *Graph) AddEdge(v1, w1 int) {
	g.adj[v1] = append(g.adj[v1], w1)
}

// String representation
func (g *Graph) String() string {
	return fmt.Sprintf("v=%v, adj=%v", g.v, g.adj)
}
