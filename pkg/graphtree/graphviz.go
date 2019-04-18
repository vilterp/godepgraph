package graphtree

import (
	"fmt"
	"strings"

	"github.com/emicklei/dot"
)

type GraphvizOpts struct {
	Horizontal bool
	NoEdges    bool
}

type NodesByName map[string]dot.Node

func addNode(g *dot.Graph, path string, node *PkgNode, nodesByName NodesByName, depth int) {
	var sg *dot.Graph
	if len(node.Children) > 0 {
		sg = g.Subgraph(node.Name, dot.ClusterOption{})
	} else {
		sg = g
	}
	g.Attr("style", "filled")
	g.Attr("color", colorForDepth(byte(depth), 255)) // TODO: darker as more deeply nested?
	if depth > 0 {
		outNode := sg.Node(node.Name).
			Box().
			Attr("fillcolor", "#afeeee").
			Attr("style", "filled").
			Attr("id", path)
		nodesByName[path] = outNode
	}
	for _, child := range node.OrderedChildren() {
		addNode(sg, path+"/"+child.Name, child, nodesByName, depth+1)
	}
}

type color struct {
	r byte
	g byte
	b byte
	a byte
}

// uugggghhhh
func hexFmt(b byte) string {
	if b < 16 {
		return fmt.Sprintf("0%x", b)
	}
	return fmt.Sprintf("%x", b)
}

func (c color) String() string {
	return fmt.Sprintf("#%s%s%s%s", hexFmt(c.r), hexFmt(c.g), hexFmt(c.b), hexFmt(c.a))
}

func colorForDepth(depth byte, maxDepth byte) color {
	return color{
		r: 0,
		g: 0,
		b: 0,
		a: depth * 10,
	}
}

func MakeGraph(tree *PkgNode, opts GraphvizOpts) *dot.Graph {
	g := dot.NewGraph(dot.Directed)
	if opts.Horizontal {
		g.Attr("rankdir", "LR")
	}
	g.Attr("splines", "ortho")
	g.Attr("nodesep", "0.4")
	g.Attr("ranksep", "0.8")

	nodesByName := map[string]dot.Node{}
	addNode(g, "", tree, nodesByName, 0)

	if opts.NoEdges {
		return g
	}

	edges := tree.getEdges("")
	for _, edge := range edges {
		if strings.HasPrefix(edge.from, "root") {
			panic("from already has root prefix")
		}
		from, ok := nodesByName[edge.from]
		if !ok {
			panic(fmt.Sprintf("can't find from node %s", edge.from))
		}
		if strings.HasPrefix(edge.to, "root") {
			panic("to already has root prefix")
		}
		to, ok := nodesByName[edge.to]
		if !ok {
			panic(fmt.Sprintf("can't find to node %s", edge.to))
		}
		g.Edge(from, to).Attr("id", fmt.Sprintf("%s->%s", edge.from, edge.to))
	}

	return g
}

func myUnusedFunc() {}
