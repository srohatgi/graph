package graph

// visit is a user defined function that is passed the vertex id
type visit func(int) error

// sort orders the vertices in order of dependencies
func sort(g *graph) []int {
	// count neighbours that point to a given vertex
	neighbours := make(map[int]int, g.vertices())

	for v := 0; v < g.vertices(); v++ {
		neighbours[v] = 0
		for w := 0; w < g.vertices(); w++ {
			if w == v {
				continue
			}
			for _, t := range g.adjascent(w) {
				if t == v {
					neighbours[v]++
				}
			}
		}
	}

	// initialize an empty list of sorted vertices
	sorted := []int{}

	for len(neighbours) != 0 {
		toRemove := -1
		for v, n := range neighbours {
			if n == 0 {
				toRemove = v
				break
			}
		}

		for _, w := range g.adjascent(toRemove) {
			neighbours[w]--
		}

		delete(neighbours, toRemove)
		sorted = append(sorted, toRemove)
	}

	return sorted
}

// dfs visits all nodes in the graph
func dfs(g *graph, visitor visit) {
	visited := make([]bool, g.vertices())

	var dfsInner func(w int) error

	dfsInner = func(w int) error {
		var err error
		for _, w := range g.adjascent(w) {
			err = dfsInner(w)
			if err != nil {
				break
			}
		}
		if err != nil {
			return err
		}
		if visited[w] {
			return nil
		}
		err = visitor(w)
		visited[w] = true
		return err
	}

	for v := 0; v < g.vertices(); v++ {
		dfsInner(v)
	}
}
