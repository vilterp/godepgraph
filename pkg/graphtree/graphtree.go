package graphtree

import (
	"fmt"
	"strings"
)

type PkgNode struct {
	Name     string
	Imports  []string
	Children map[string]*PkgNode // segment => child node
	Size     int
}

func (n *PkgNode) Print() {
	n.printTreeHelper(0)
}

func (n *PkgNode) printTreeHelper(level int) {
	fmt.Printf("%s%s %d\n", strings.Repeat("  ", level), n.Name, len(n.Imports))

	for _, child := range n.Children {
		child.printTreeHelper(level + 1)
	}
}

type edge struct {
	from string
	to   string
}

func (n *PkgNode) getEdges(path string) []edge {
	return n.getEdgesHelper(path, nil)
}

func (n *PkgNode) getEdgesHelper(path string, edges []edge) []edge {
	for _, imp := range n.Imports {
		edges = append(edges, edge{
			from: path,
			to:   "/" + imp,
		})
	}
	for _, c := range n.Children {
		edges = c.getEdgesHelper(path+"/"+c.Name, edges)
	}
	return edges
}
