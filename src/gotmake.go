// Copyright 2009 Dimiter Stanev, malkia@gmail.com.
// Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"go/parser"
	"go/ast"
	"strconv"
	"path"
  "sort"
	"flag"
	"bytes"
	"go/printer"
	stringslice "./gotgo/slice(string)"
)

var ignoreInstalled = flag.Bool("ignore-installed", true,
	"don't require that installed gotgo packages be present")

var noGotgo = flag.Bool("without-gotgo", true,
	"don't require that gotgo be installed")

func getImports(filename string) (pkgname string,
	imports map[string]string, names map[string]string, error os.Error) {
	source, error := ioutil.ReadFile(filename)
	if error != nil {
		return
	}
	file, error := parser.ParseFile(filename, source, nil, parser.ImportsOnly)
	if error != nil {
		return
	}
	dir, _ := path.Split(filename)
	pkgname = file.Name.Name()
	for _, importDecl := range file.Decls {
		importDecl, ok := importDecl.(*ast.GenDecl)
		if ok {
			for _, importSpec := range importDecl.Specs {
				importSpec, ok := importSpec.(*ast.ImportSpec)
				if ok {
					importPath, _ := strconv.Unquote(string(importSpec.Path.Value))
					if len(importPath) > 0 {
						if imports == nil {
							imports = make(map[string]string)
							names = make(map[string]string)
						}
						if importPath[0] == '.' || strings.Index(importPath, "(") != -1 {
							nicepath := path.Join(dir, path.Clean(importPath))
							imports[nicepath] = importPath
							//names[path.Join(dir,path.Clean(importSpec.Name.String()))]=importPath
							names[importSpec.Name.String()]=nicepath
						}
					}
				}
			}
		}
	}
	return
}

func cleanbinname(f string) string {
	if f[0:4] == "src/" { return "bin/"+f[4:] }
	return f
}

func cleanpkgname(f string) string {
	if startswith(f, "pkg/") { return f[4:] }
	return f
}

type maker struct {
}
func (maker) VisitDir(string, *os.Dir) bool { return true }
func (maker) VisitFile(f string, stat *os.Dir) {
	if !stat.IsRegular() { return } // it's something weird...
	if endswith(f, ".gotgo.go") {
		fmt.Printf("# ignoring %s, since it's a generated file\n", f)
	} else if endswith(f, ".got.go") {
		fmt.Printf("# ignoring %s, since it's a generated file\n", f)
	} else if endswith(f, ".go") {
		deps := []string{f}
		pname, imports, names, err := getImports(f)
		if err != nil {
			fmt.Println("# error: ", err)
		}
		for i,ii := range imports {
			// fmt.Printf("# %s imports %s\n", f, i)
			createGofile(i, ii, names)
			deps = stringslice.Append(deps, i + ".$(O)")
		}
    sort.SortStrings(deps[1:]) // alphebatize all deps but first
		basename := f[0:len(f)-3]
		objname := basename+".$(O)"
		if pname == "main" {
			fmt.Printf("%s: %s\n\t@mkdir -p bin\n\t$(LD) -o $@ $<\n", cleanbinname(basename), objname)
			if cleanbinname(basename)[0:4] == "bin/" {
				installname := fmt.Sprintf("$(bindir)/%s",
					cleanbinname(basename)[4:])
				fmt.Printf("%s: %s\n\tcp $< $@\n", installname, cleanbinname(basename))
				installbins += " " + installname
			}
		}
		if isDone(objname) {
			// We've already got a rule for this object file, so just quit now.
			return
		}
		fmt.Print(objname+":")
		done(objname)
		for _,d := range deps {
			fmt.Print(" "+d)
		}
		fmt.Print("\n")
		if startswith(basename, "pkg/") && strings.Index(basename,"(") == -1 &&
			!endswith(basename,"got.go") {
			// we want to install this as a package
			installname := fmt.Sprintf("$(pkgdir)/%s.a", basename[4:])
			dir,_ := path.Split(installname)
			fmt.Printf("%s.a: %s\n\tgopack grc $@ $<\n", basename, objname)
			fmt.Printf("%s: %s.a\n\tmkdir -p %s\n\tcp $< $@\n\n",
				installname, basename, dir)
			done(installname)
			installpkgs += " " + installname
		}
		fmt.Print("\n")
	} else if path.Ext(f) == ".got" {
		fmt.Printf("# found file %s to build...\n", f)
		// YUCK: the following code essentially duplicates the code in the
		// previous if case!
		deps := []string{}
		_, imports, names, err := getImports(f)
		if err != nil {
			fmt.Println("# error: ", err)
		}
		for i,ii := range imports {
			fmt.Printf("# %s depends on %s\n", f, i)
			createGofile(i, ii, names)
			deps = stringslice.Append(deps, i + ".$(O)")
		}
    sort.SortStrings(deps) // alphebatize all these deps
		if isDone(f+"go") {
			// Quit early if we've already got a rule for making the .gotgo file.
			return
		}
		fmt.Printf("%s: %s", f+"go", f)
		done(f+"go")
		for _,d := range deps {
			fmt.Print(" "+d)
		}
		fmt.Print("\n")
		if startswith(f, "pkg/") {
			// we want to install this as a got package!
			installname := fmt.Sprintf("$(pkgdir)/%s", f[4:]+"go")
			dir,_ := path.Split(installname)
			fmt.Printf("%s: %s\n\tmkdir -p %s\n\tcp $< $@\n\n",installname,f+"go",dir)
			done(installname)
			installpkgs += " " + installname
		}
		fmt.Print("\n")
	}
}

func createGofile(sourcePath, importPath string, names map[string]string) {
	switch n := strings.Index(sourcePath, "("); n {
	case -1:
		// ignore non-templated import...
	default:
		// Begin by scanning the type parameters in the import
		types, imps := ParseImport(sourcePath)
		//fmt.Printf("# I see types %v and imports %v\n", types, imps)
		basename := sourcePath[0:n]
		gotname := basename + ".got"
		gotgoname := gotname + "go"
		goname := sourcePath + ".go"
		nip := strings.Index(importPath, "(")
		pkggotname := path.Join(pkgdir,path.Clean(importPath[0:nip]+".gotgo"))

		if isDone(goname) {
			// We already know how to build goname...
			return
		} else if *noGotgo {
			// We don't want to rely on gotgo existing.
			fmt.Printf("# File %s is generated from a got template.\n", goname)
			return
		}
		fmt.Printf("ifneq ($(strip $(shell which gotgo)),)\n")
		defer fmt.Printf("endif\n")
		if fileexists(gotname) {
			fmt.Printf("%s: %s\n\t$<", goname, gotgoname)
			done(goname)
		} else if fileexists(pkggotname) {
			if *ignoreInstalled {
				// We don't want the makefile to depend on installed packages.
				fmt.Printf("# we assume %s is present...\n", goname)
				return
			}
			fmt.Printf("# looks like we require %s as installed package...\n",
				gotname)
			dir,_ := path.Split(goname)
			fmt.Printf("%s: $(pkgdir)/%s\n\tmkdir -p %s\n\t$<",
				goname, importPath[0:nip]+".gotgo", dir)
			done(goname)
		} else {
			fmt.Printf("# I don't know how to make %s from %s or %s\n",
				goname, gotname, pkggotname)
			return
		}
		// Now we want to figure out how to modify the import path The
		// import paths must be written relative to sourcePath, but we
		// only know them relative to the Makefile position.
		relnames := make(map[string]string)
		// ds is the set of directories we might care to strip...
		ds := stringslice.Reverse(dirs(path.Clean(goname)))
		for k,v := range names {
			gotmismatch := false
			for _,d := range ds {
				if !gotmismatch && len(v) > len(d) && v[0:len(d)+1] == d+"/" {
					v = v[len(d)+1:]
				} else {
					v = "../" + v
				}
			}
			if v[0] != '.' { v = "./" + v }	
			relnames[k] = v
		}
		importsdone := make(map[string]bool)
		for _, i := range imps {
			_, alreadydone := importsdone[i]
			if alreadydone { continue }
			n, ok := relnames[i]
			if !ok {
				fmt.Print(" --import 'import "+strconv.Quote(i)+"'") // FIXME
			} else {
				importsdone[i] = true
				fmt.Print(" --import 'import "+i+" "+strconv.Quote(n)+"'")
			}
		}
		for _, a := range types {
			fmt.Printf(" '%s'", a)
		}
		fmt.Printf(" > \"$@\"\n")
	}
	return
}

type seeker struct {
}
func (seeker) VisitDir(string, *os.Dir) bool { return true }
func (seeker) VisitFile(f string, _ *os.Dir) {
	if endswith(f, ".got.go") {
		// don't do anything, since this is a temporary file...
	} else if strings.Index(f,"(") != -1 {
		// don't do anything, since it's a templated file...
	} else if endswith(f, ".go") {
		pname, _, _, _ := getImports(f)
		basename := f[0:len(f)-3]
		if pname == "main" && len(f) > 4 && f[0:4] != "test" &&
			!endswith(basename, ".gotgo") {
			mybinfiles += " " + cleanbinname(basename)
		} else if startswith(f, "pkg/") {
			mypackages += " " + basename + ".a"
		}
	} else if endswith(f, ".got") {
		if startswith(f, "pkg/") {
			mypackages += " " + f + "go"
		}
	}
}

type void struct{}
var alreadyDone = make(map[string]void)
func isDone(targ string) bool {
	_, ok := alreadyDone[targ]
	return ok
}
func done(targ string) { alreadyDone[targ] = void{} }

var mybinfiles = ""
var mypackages = ""
var installbins = ""
var installpkgs = ""
var pkgdir = path.Join(os.Getenv("GOROOT"),"pkg",
	os.Getenv("GOOS")+"_"+os.Getenv("GOARCH"))

func main() {
	flag.Parse()
	path.Walk(".", seeker{}, nil)
	fmt.Printf(`

include $(GOROOT)/src/Make.$(GOARCH)

binaries: %s
packages: %s

ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))

.PHONY: test binaries packages install installbins installpkgs $(EXTRAPHONY)
.SUFFIXES: .$(O) .go .got .gotgo $(EXTRASUFFIXES)

`, mybinfiles, mypackages)
	fmt.Print(".go.$(O):\n\tcd `dirname \"$<\"`; $(GC) `basename \"$<\"`\n")
	if fileexists("gotgo.go") {
		fmt.Print(".got.gotgo:\n\t./gotgo \"$<\"\n\n")
	} else {
		fmt.Print(".got.gotgo:\n\tgotgo \"$<\"\n\n")
	}
	path.Walk(".", maker{}, nil)
	fmt.Printf("installbins: %s\n", installbins)
	fmt.Printf("installpkgs: %s\n", installpkgs)
}

func fileexists(f string) bool {
	x, err := os.Stat(f)
	return err == nil && x.IsRegular()
}

func endswith(s, end string) bool {
	return len(s) >= len(end) && s[len(s) - len(end):] == end
}

func startswith(s, st string) bool {
	return len(s) >= len(st) && s[0:len(st)] == st
}

func taketo(s, sep string) string {
	n := strings.Index(s, sep)
	if n == -1 { return "" }
	return s[0:n]
}

func dirs(s string) (out []string) {
	s,_ = path.Split(s) // drop filename
	if s == "" { return out }
	s = s[0:len(s)-1]
	for {
		d,ld := path.Split(s)
		out = stringslice.Append(out, ld)
		if d == "" { return out }
		s = d[0:len(d)-1]
	}
	return out
}

// This is a convenience type for finding the imports of a file.
type getimps struct {
	imps *[]string
}
func (g getimps) Visit(x interface{}) ast.Visitor {
	switch v := x.(type) {
	case *ast.SelectorExpr:
		*g.imps = stringslice.Append(*g.imps, pretty(v.X))
	}
	if x != nil {
		return g
	}
	return nil
}

// This parses an import statement, such as "./foo/bar(int,baz.T(foo.T))",
// returning a list of types and a list of imports that are needed
// (e.g. baz and foo in the above example).
func ParseImport(s string) (types, imports []string) {
	// First, I want to cut off any preliminary directories, so the
	// import should look like a function call.
	n := strings.Index(s, "(")
	if n < 1 { return }
	start := 0
	for i := n-1; i>=0; i-- {
		if s[i] == '/' {
			start = i+1
			break
		}
	}
	s = s[start:]
	// Now we just need to parse the apparent function call...
	x, _ := parser.ParseExpr(s, s, nil)
	is := []string{}
	callexpr, ok := x.(*ast.CallExpr)
	if !ok { return } // FIXME: need error handling?
	for _, texpr := range callexpr.Args {
		types = stringslice.Append(types, pretty(texpr))
		ast.Walk(getimps{&is}, texpr)
	}
	return types, is
}

func pretty(expr interface{}) string {
	b := bytes.NewBufferString("")
	printer.Fprint(b, expr)
	return b.String()
}
