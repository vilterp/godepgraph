package graphtree

import (
	"fmt"
	"go/build"
	"strings"
)

type PkgNode struct {
	Name     string
	Imports  []string
	Children map[string]*PkgNode // segment => child node
}

func (n *PkgNode) Print(pkgs map[string]*build.Package) {
	n.printTreeHelper(pkgs, 0)
}

func (n *PkgNode) printTreeHelper(pkgs map[string]*build.Package, level int) {
	fmt.Printf("%s%s %s\n", strings.Repeat("  ", level), n.Name, n.Imports)

	for _, child := range n.Children {
		child.printTreeHelper(pkgs, level+1)
	}
}

type edge struct {
	from string
	to   string
}

func (n *PkgNode) getEdges(path string) []*edge {
	return n.getEdgesHelper(path, nil)
}

func (n *PkgNode) getEdgesHelper(path string, edges []*edge) []*edge {
	for _, imp := range n.Imports {
		edges = append(edges, &edge{
			from: path,
			to:   imp,
		})
	}
	for _, c := range n.Children {
		edges = c.getEdgesHelper(path+"/"+c.Name, edges)
	}
	return edges
}
