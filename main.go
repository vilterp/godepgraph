package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/build"
	"log"
	"os"
	"strings"

	"github.com/vilterp/godepgraph/pkg/graphtree"
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
	path           = flag.String("path", "", "path to go into") // TODO: better explanation

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

	b := graphtree.NewBuilder(graphtree.Opts{
		MaxLevel:        *maxLevel,
		WithTests:       *withTests,
		OnlyPrefixes:    onlyPrefixes,
		IgnoreVendor:    *ignoreVendor,
		IgnoreStdlib:    *ignoreStdlib,
		BuildContext:    buildContext,
		WithGoRoot:      *withGoroot,
		Ignored:         ignored,
		IgnoredPrefixes: ignoredPrefixes,
	})
	for _, a := range args {
		if err := b.ProcessPackage(cwd, a, 0, "", *stopOnError); err != nil {
			log.Fatal(err)
		}
	}

	pkgTree := b.GetTree()

	if len(*path) > 0 {
		pkgTree = pkgTree.GetChild(strings.Split(*path, "/"))
		if pkgTree == nil {
			log.Fatal("error getting child at path")
		}
	}

	if *printJson {
		_, err := json.MarshalIndent(pkgTree, "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		//if _, err := os.Stdout.Write(bytes); err != nil {
		//	log.Fatal(err)
		//}
	} else {
		g := graphtree.MakeGraph(pkgTree, graphtree.GraphvizOpts{
			Horizontal: true,
		})
		fmt.Println(g.String())
	}
}

func pkgURL(pkgName string) string {
	return "https://godoc.org/" + pkgName
}

func debug(args ...interface{}) {
	fmt.Fprintln(os.Stderr, args...)
}

func debugf(s string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, s, args...)
}
