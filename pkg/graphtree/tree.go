package graphtree

import (
	"go/build"
	"strings"
)

func (b *Builder) GetTree() *PkgNode {
	root := &PkgNode{
		Name:     "root",
		Children: map[string]*PkgNode{},
	}
	for name, pkg := range b.pkgs {
		root.insertAtPath(b, strings.Split(name, "/"), pkg)
	}
	return root
}

func (n *PkgNode) size() int {
	ret := 1
	for _, c := range n.Children {
		ret += c.size()
	}
	return ret
}

func (n *PkgNode) insertAtPath(b *Builder, path []string, pkg *build.Package) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		if n.Children[path[0]] != nil {
			n.Children[path[0]].Imports = b.importsForPkg(pkg)
			return
		}
		n.Children[path[0]] = &PkgNode{
			Name:     path[0],
			Imports:  b.importsForPkg(pkg),
			Children: map[string]*PkgNode{},
		}
		return
	}
	child := n.Children[path[0]]
	if child == nil {
		child = &PkgNode{
			Name:     path[0],
			Imports:  b.importsForPkg(pkg),
			Children: map[string]*PkgNode{},
		}
		n.Children[path[0]] = child
	}
	child.insertAtPath(b, path[1:], pkg)
}

func (n *PkgNode) GetChild(path []string) *PkgNode {
	return n.getChildHelp(path, path)
}

func (n *PkgNode) getChildHelp(path []string, wholePath []string) *PkgNode {
	if len(path) == 0 {
		return n.rewriteImports(strings.Join(wholePath, "/"))
	}
	nextChild := n.Children[path[0]]
	if nextChild == nil {
		return nil
	}
	return nextChild.getChildHelp(path[1:], wholePath)
}

func (n *PkgNode) rewriteImports(removePrefix string) *PkgNode {
	out := &PkgNode{
		Name:     n.Name,
		Children: map[string]*PkgNode{},
	}
	for _, imp := range n.Imports {
		out.Imports = append(out.Imports, imp[len(removePrefix)+1:])
	}
	for name, child := range n.Children {
		out.Children[name] = child.rewriteImports(removePrefix)
	}
	return out
}
