package graph

// Visit is a user defined function that is passed the vertex id
type Visit func(int) error

// Sort orders the vertices in order of dependencies
func Sort(g *Graph) []int {
	// count neighbours that point to a given vertex
	neighbours := make(map[int]int, g.Vertices())

	for v := 0; v < g.Vertices(); v++ {
		neighbours[v] = 0
		for w := 0; w < g.Vertices(); w++ {
			if w == v {
				continue
			}
			for _, t := range g.Adjascent(w) {
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

		for _, w := range g.Adjascent(toRemove) {
			neighbours[w]--
		}

		delete(neighbours, toRemove)
		sorted = append(sorted, toRemove)
	}

	logger("sorted", sorted)
	return sorted
}

// DFS visits all nodes in the graph
func DFS(g *Graph, visit Visit) {
	visited := make([]bool, g.Vertices())

	var dfs func(w int) error

	dfs = func(w int) error {
		var err error
		for _, w := range g.Adjascent(w) {
			err = dfs(w)
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
		err = visit(w)
		visited[w] = true
		return err
	}

	for v := 0; v < g.Vertices(); v++ {
		dfs(v)
	}
}
