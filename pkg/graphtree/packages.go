package graphtree

import (
	"fmt"
	"go/build"
	"strings"
)

type Builder struct {
	pkgs map[string]*build.Package
	ids  map[string]string
	opts Opts
}

type Opts struct {
	WithTests       bool
	OnlyPrefixes    []string
	IgnoreVendor    bool
	IgnoreStdlib    bool
	MaxLevel        int
	BuildContext    build.Context
	WithGoRoot      bool
	Ignored         map[string]bool
	IgnoredPrefixes []string
}

func NewBuilder(opts Opts) *Builder {
	return &Builder{
		pkgs: map[string]*build.Package{},
		ids:  map[string]string{},
		opts: opts,
	}
}

// TODO: somthing is wrong here
//   returning imports for treenodes that have no imports
func (b *Builder) importsForPkg(pkg *build.Package) []string {
	if pkg == nil {
		return nil
	}

	var out []string
	for _, imp := range b.getImports(pkg) {
		impPkg := b.pkgs[imp]
		if impPkg == nil || b.isIgnored(impPkg) {
			continue
		}
		out = append(out, imp)
	}
	return out
}

func (b *Builder) getImports(pkg *build.Package) []string {
	allImports := pkg.Imports
	if b.opts.WithTests {
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

func hasPrefixes(s string, prefixes []string) bool {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func (b *Builder) isIgnored(pkg *build.Package) bool {
	if len(b.opts.OnlyPrefixes) > 0 && !hasPrefixes(normalizeVendor(pkg.ImportPath), b.opts.OnlyPrefixes) {
		return true
	}

	if b.opts.IgnoreVendor && isVendored(pkg.ImportPath) {
		return true
	}
	return b.opts.Ignored[normalizeVendor(pkg.ImportPath)] || (pkg.Goroot && b.opts.IgnoreStdlib) || hasPrefixes(normalizeVendor(pkg.ImportPath), b.opts.IgnoredPrefixes)
}

func isVendored(path string) bool {
	return strings.Contains(path, "/vendor/")
}

func normalizeVendor(path string) string {
	pieces := strings.Split(path, "vendor/")
	return pieces[len(pieces)-1]
}

func (b *Builder) ProcessPackage(
	root string, pkgName string, level int, importedBy string, stopOnError bool,
) error {
	if level++; level > b.opts.MaxLevel {
		return nil
	}
	if b.opts.Ignored[pkgName] {
		return nil
	}

	pkg, err := b.opts.BuildContext.Import(pkgName, root, 0)
	if err != nil {
		if stopOnError {
			return fmt.Errorf("failed to import %s (imported at level %d by %s): %s", pkgName, level, importedBy, err)
		} else {
			// TODO: mark the package so that it is rendered with a different color
		}
	}

	if b.isIgnored(pkg) {
		return nil
	}

	b.pkgs[normalizeVendor(pkg.ImportPath)] = pkg

	// Don't worry about dependencies for stdlib packages
	if pkg.Goroot && !b.opts.WithGoRoot {
		return nil
	}

	for _, imp := range b.getImports(pkg) {
		if _, ok := b.pkgs[imp]; !ok {
			if err := b.ProcessPackage(pkg.Dir, imp, level, pkgName, stopOnError); err != nil {
				return err
			}
		}
	}
	return nil
}

func (b *Builder) GetPackages() map[string]*build.Package {
	return b.pkgs
}
