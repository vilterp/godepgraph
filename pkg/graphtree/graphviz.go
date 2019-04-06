package graphtree

import (
	"fmt"
	"strings"

	"github.com/emicklei/dot"
)

type GraphvizOpts struct {
	Horizontal bool
}

type NodesByName map[string]dot.Node

func addNode(g *dot.Graph, path string, node *PkgNode, nodesByName NodesByName, depth int) {
	var sg *dot.Graph
	if len(node.Children) > 0 {
		sg = g.Subgraph(path, dot.ClusterOption{})
	} else {
		sg = g
	}
	g.Attr("style", "filled")
	g.Attr("color", colorForDepth(byte(depth), 255)) // TODO: darker as more deeply nested?
	outNode := sg.Node(path).Box()
	nodesByName[path] = outNode
	for _, child := range node.Children {
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
	edges := tree.getEdges("")

	g := dot.NewGraph(dot.Directed)
	if opts.Horizontal {
		g.Attr("rankdir", "LR")
	}
	g.Attr("splines", "ortho")
	g.Attr("nodesep", "0.4")
	g.Attr("ranksep", "0.8")

	nodesByName := map[string]dot.Node{}
	addNode(g, "", tree, nodesByName, 0)

	//var nbn []string
	//for n := range nodesByName {
	//	nbn = append(nbn, n)
	//}
	//sort.Strings(nbn)
	//for _, n := range nbn {
	//	fmt.Println(n)
	//}

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
		g.Edge(from, to)
	}

	//node [shape="box",style="rounded,filled"]
	//edge [arrowsize="0.5"]

	//for _, pkgName := range pkgKeys {
	//	pkg := pkgs[pkgName]
	//	pkgId := getId(ids, pkgName)
	//
	//	if isIgnored(pkg) {
	//		continue
	//	}
	//
	//	var color string
	//	switch {
	//	case pkg.Goroot:
	//		color = "palegreen"
	//	case len(pkg.CgoFiles) > 0:
	//		color = "darkgoldenrod1"
	//	case isVendored(pkg.ImportPath):
	//		color = "palegoldenrod"
	//	default:
	//		color = "paleturquoise"
	//	}
	//
	//	fmt.Printf("%s [label=\"%s\" color=\"%s\" URL=\"%s\" target=\"_blank\"];\n", pkgId, pkgName, color, pkgURL(pkgName))
	//
	//	// Don't render imports from packages in Goroot
	//	if pkg.Goroot && !*withGoroot {
	//		continue
	//	}
	//
	//	for _, imp := range getImports(pkg) {
	//		impPkg := pkgs[imp]
	//		if impPkg == nil || isIgnored(impPkg) {
	//			continue
	//		}
	//
	//		impId := getId(ids, imp)
	//		fmt.Printf("%s -> %s;\n", pkgId, impId)
	//	}
	//}

	return g
}
