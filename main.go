package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"sort"
	"strings"
)

var (
	ignored = map[string]bool{
		"C": true,
	}
	ignoredPrefixes []string
	onlyPrefixes    []string

	ignoreStdlib   = flag.Bool("nostdlib", false, "ignore packages in the Go standard library")
	ignoreVendor   = flag.Bool("novendor", false, "ignore packages in the vendor directory")
	stopOnError    = flag.Bool("stoponerror", true, "stop on package import errors")
	withGoroot     = flag.Bool("withgoroot", false, "show dependencies of packages in the Go standard library")
	ignorePrefixes = flag.String("ignoreprefixes", "", "a comma-separated list of prefixes to ignore")
	ignorePackages = flag.String("ignorepackages", "", "a comma-separated list of packages to ignore")
	onlyPrefix     = flag.String("onlyprefixes", "", "a comma-separated list of prefixes to include")
	tagList        = flag.String("tags", "", "a comma-separated list of build tags to consider satisfied during the build")
	horizontal     = flag.Bool("horizontal", false, "lay out the dependency graph horizontally instead of vertically")
	withTests      = flag.Bool("withtests", false, "include test packages")
	maxLevel       = flag.Int("maxlevel", 256, "max level of go dependency graph")
	printJson      = flag.Bool("jsontree", false, "print tree of packages as JSON")

	buildTags    []string
	buildContext = build.Default
)

func init() {
	flag.BoolVar(ignoreStdlib, "s", false, "(alias for -nostdlib) ignore packages in the Go standard library")
	flag.StringVar(ignorePrefixes, "p", "", "(alias for -ignoreprefixes) a comma-separated list of prefixes to ignore")
	flag.StringVar(ignorePackages, "i", "", "(alias for -ignorepackages) a comma-separated list of packages to ignore")
	flag.StringVar(onlyPrefix, "o", "", "(alias for -onlyprefixes) a comma-separated list of prefixes to include")
	flag.BoolVar(withTests, "t", false, "(alias for -withtests) include test packages")
	flag.IntVar(maxLevel, "l", 256, "(alias for -maxlevel) maximum level of the go dependency graph")
	flag.BoolVar(withGoroot, "d", false, "(alias for -withgoroot) show dependencies of packages in the Go standard library")
}

func main() {
	flag.Parse()

	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("need one package name to process")
	}

	if *ignorePrefixes != "" {
		ignoredPrefixes = strings.Split(*ignorePrefixes, ",")
	}
	if *onlyPrefix != "" {
		onlyPrefixes = strings.Split(*onlyPrefix, ",")
	}
	if *ignorePackages != "" {
		for _, p := range strings.Split(*ignorePackages, ",") {
			ignored[p] = true
		}
	}
	if *tagList != "" {
		buildTags = strings.Split(*tagList, ",")
	}
	buildContext.BuildTags = buildTags

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("failed to get cwd: %s", err)
	}

	pkgs := map[string]*build.Package{}
	ids := map[string]string{}
	for _, a := range args {
		if err := processPackage(pkgs, ids, cwd, a, 0, "", *stopOnError); err != nil {
			log.Fatal(err)
		}
	}

	// sort packages
	pkgKeys := []string{}
	for k := range pkgs {
		pkgKeys = append(pkgKeys, k)
	}
	sort.Strings(pkgKeys)

	if *printJson {
		pkgTree := buildPkgTree(pkgs)
		bytes, err := json.MarshalIndent(pkgTree, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(os.Stdout.Write(bytes))
	} else {
		printGraph(pkgs, ids, pkgKeys)
	}
}

func printGraph(pkgs map[string]*build.Package, ids map[string]string, pkgKeys []string) {
	fmt.Println("digraph godep {")
	if *horizontal {
		fmt.Println(`rankdir="LR"`)
	}
	fmt.Print(`splines=ortho
nodesep=0.4
ranksep=0.8
node [shape="box",style="rounded,filled"]
edge [arrowsize="0.5"]
`)

	for _, pkgName := range pkgKeys {
		pkg := pkgs[pkgName]
		pkgId := getId(ids, pkgName)

		if isIgnored(pkg) {
			continue
		}

		var color string
		switch {
		case pkg.Goroot:
			color = "palegreen"
		case len(pkg.CgoFiles) > 0:
			color = "darkgoldenrod1"
		case isVendored(pkg.ImportPath):
			color = "palegoldenrod"
		default:
			color = "paleturquoise"
		}

		fmt.Printf("%s [label=\"%s\" color=\"%s\" URL=\"%s\" target=\"_blank\"];\n", pkgId, pkgName, color, pkgURL(pkgName))

		// Don't render imports from packages in Goroot
		if pkg.Goroot && !*withGoroot {
			continue
		}

		for _, imp := range getImports(pkg) {
			impPkg := pkgs[imp]
			if impPkg == nil || isIgnored(impPkg) {
				continue
			}

			impId := getId(ids, imp)
			fmt.Printf("%s -> %s;\n", pkgId, impId)
		}
	}

	fmt.Println("}")
}

// todo: marshal as json
type pkgNode struct {
	Name     string
	Imports  []string
	Children map[string]*pkgNode // segment => child node
}

func buildPkgTree(pkgs map[string]*build.Package) *pkgNode {
	root := &pkgNode{
		Name:     "root",
		Children: map[string]*pkgNode{},
	}
	for name, pkg := range pkgs {
		root.insertAtPath(strings.Split(name, "/"), pkg, pkgs)
	}
	return root
}

func (n *pkgNode) print(pkgs map[string]*build.Package) {
	n.printTreeHelper(pkgs, 0)
}

func (n *pkgNode) printTreeHelper(pkgs map[string]*build.Package, level int) {
	fmt.Printf("%s%s %s\n", strings.Repeat("  ", level), n.Name, n.Imports)

	for _, child := range n.Children {
		child.printTreeHelper(pkgs, level+1)
	}
}

func (n *pkgNode) insertAtPath(path []string, pkg *build.Package, pkgs map[string]*build.Package) {
	if len(path) == 0 {
		return
	}
	if len(path) == 1 {
		n.Children[path[0]] = &pkgNode{
			Name:     path[0],
			Imports:  importsForPkg(pkg, pkgs),
			Children: map[string]*pkgNode{},
		}
		return
	}
	child := n.Children[path[0]]
	if child == nil {
		child = &pkgNode{
			Name:     path[0],
			Imports:  importsForPkg(pkg, pkgs),
			Children: map[string]*pkgNode{},
		}
		n.Children[path[0]] = child
	}
	child.insertAtPath(path[1:], pkg, pkgs)
}

func importsForPkg(pkg *build.Package, pkgs map[string]*build.Package) []string {
	if pkg == nil {
		return nil
	}

	var out []string
	for _, imp := range getImports(pkg) {
		impPkg := pkgs[imp]
		if impPkg == nil || isIgnored(impPkg) {
			continue
		}
		out = append(out, imp)
	}
	return out
}

func pkgURL(pkgName string) string {
	return "https://godoc.org/" + pkgName
}

func processPackage(
	pkgs map[string]*build.Package, ids map[string]string, root string, pkgName string,
	level int, importedBy string, stopOnError bool,
) error {
	if level++; level > *maxLevel {
		return nil
	}
	if ignored[pkgName] {
		return nil
	}

	pkg, err := buildContext.Import(pkgName, root, 0)
	if err != nil {
		if stopOnError {
			return fmt.Errorf("failed to import %s (imported at level %d by %s): %s", pkgName, level, importedBy, err)
		} else {
			// TODO: mark the package so that it is rendered with a different color
		}
	}

	if isIgnored(pkg) {
		return nil
	}

	pkgs[normalizeVendor(pkg.ImportPath)] = pkg

	// Don't worry about dependencies for stdlib packages
	if pkg.Goroot && !*withGoroot {
		return nil
	}

	for _, imp := range getImports(pkg) {
		if _, ok := pkgs[imp]; !ok {
			if err := processPackage(pkgs, ids, pkg.Dir, imp, level, pkgName, stopOnError); err != nil {
				return err
			}
		}
	}
	return nil
}

func getImports(pkg *build.Package) []string {
	allImports := pkg.Imports
	if *withTests {
		allImports = append(allImports, pkg.TestImports...)
		allImports = append(allImports, pkg.XTestImports...)
	}
	var imports []string
	found := make(map[string]struct{})
	for _, imp := range allImports {
		if imp == normalizeVendor(pkg.ImportPath) {
			// Don't draw a self-reference when foo_test depends on foo.
			continue
		}
		if _, ok := found[imp]; ok {
			continue
		}
		found[imp] = struct{}{}
		imports = append(imports, imp)
	}
	return imports
}

func deriveNodeID(packageName string) string {
	//TODO: improve implementation?
	id := "\"" + packageName + "\""
	return id
}

func getId(ids map[string]string, name string) string {
	id, ok := ids[name]
	if !ok {
		id = deriveNodeID(name)
		ids[name] = id
	}
	return id
}

func hasPrefixes(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func isIgnored(pkg *build.Package) bool {
	if len(onlyPrefixes) > 0 && !hasPrefixes(normalizeVendor(pkg.ImportPath), onlyPrefixes) {
		return true
	}

	if *ignoreVendor && isVendored(pkg.ImportPath) {
		return true
	}
	return ignored[normalizeVendor(pkg.ImportPath)] || (pkg.Goroot && *ignoreStdlib) || hasPrefixes(normalizeVendor(pkg.ImportPath), ignoredPrefixes)
}

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func debugf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
}

func isVendored(path string) bool {
	return strings.Contains(path, "/vendor/")
}

func normalizeVendor(path string) string {
	pieces := strings.Split(path, "vendor/")
	return pieces[len(pieces)-1]
}
