package graphtree

import (
	"fmt"
	"sort"
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

	for _, child := range n.OrderedChildren() {
		child.printTreeHelper(level + 1)
	}
}

func (n *PkgNode) OrderedChildren() []*PkgNode {
	var names []string
	for name := range n.Children {
		names = append(names, name)
	}
	sort.Strings(names)
	var out []*PkgNode
	for _, name := range names {
		out = append(out, n.Children[name])
	}
	return out
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
	for _, c := range n.OrderedChildren() {
		edges = c.getEdgesHelper(path+"/"+c.Name, edges)
	}
	return edges
}
