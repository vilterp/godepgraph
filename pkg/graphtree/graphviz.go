package graphtree

import (
	"github.com/emicklei/dot"
)

type GraphvizOpts struct {
	Horizontal bool
}

type NodesByName map[string]dot.Node

func addNode(g *dot.Graph, path string, node *PkgNode, nodesByName NodesByName) {
	sg := g.Subgraph(path, dot.ClusterOption{})
	outNode := sg.Node(path)
	nodesByName[path] = outNode
	for _, child := range node.Children {
		addNode(sg, path+"/"+child.Name, child, nodesByName)
	}
}

func MakeGraph(tree *PkgNode, opts GraphvizOpts) *dot.Graph {
	edges := tree.getEdges(tree.Name)

	g := dot.NewGraph(dot.Directed)
	if opts.Horizontal {
		g.Attr("rankdir", "LR")
	}
	g.Attr("splines", "ortho")
	g.Attr("nodesep", "0.4")
	g.Attr("ranksep", "0.8")

	nodesByName := map[string]dot.Node{}
	addNode(g, tree.Name, tree, nodesByName)

	for _, edge := range edges {
		g.Edge(nodesByName[edge.from], nodesByName[edge.to])
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
