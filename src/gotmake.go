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
	"go/scanner"
	"strconv"
	"path"
  "sort"
	"./got/buildit"
)

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
	if f[0:4] == "pkg/" { return f[4:] }
	return f
}

type maker struct {
}
func (maker) VisitDir(string, *os.Dir) bool { return true }
func (maker) VisitFile(f string, _ *os.Dir) {
	if endswith(f, ".gotgo.go") {
		fmt.Printf("# ignoring %s, since it's a generated file\n", f)
	} else if endswith(f, ".go") {
		deps := make([]string, 1, 1000) // FIXME stupid hardcoded limit...x
		deps[0] = f
		pname, imports, names, err := getImports(f)
		if err != nil {
			fmt.Println("# error: ", err)
		}
		for i,ii := range imports {
			fmt.Printf("# %s imports %s\n", f, i)
			createGofile(i, ii, names)
			deps = deps[0:len(deps)+1]
			deps[len(deps)-1] = i + ".$(O)"
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
		fmt.Print(objname+":")
		for _,d := range deps {
			fmt.Print(" "+d)
		}
		fmt.Print("\n")
		if basename[0:4] == "pkg/" && !endswith(basename,"gotgo.go") {
			// we want to install this as a package
			installname := fmt.Sprintf("$(pkgdir)/%s", objname)
			dir,_ := path.Split(installname)
			fmt.Printf("%s: %s\n\tmkdir -p %s\n\tcp $< $@\n\n",
				installname, objname, dir)
			installpkgs += " " + installname
		}
		fmt.Print("\n")
	} else if path.Ext(f) == ".got" {
		fmt.Printf("# found file %s to build...\n", f)
		// YUCK: the following code essentially duplicates the code in the
		// previous if case!
		deps := make([]string, 1, 1000) // FIXME stupid hardcoded limit...x
		_, imports, names, err := getImports(f)
		if err != nil {
			fmt.Println("# error: ", err)
		}
		for i,ii := range imports {
			fmt.Printf("# %s depends on %s\n", f, i)
			createGofile(i, ii, names)
			deps = deps[0:len(deps)+1]
			deps[len(deps)-1] = i + ".$(O)"
		}
    sort.SortStrings(deps) // alphebatize all these deps
		fmt.Printf("%s: %s", f+"go", f)
		for _,d := range deps {
			fmt.Print(" "+d)
		}
		fmt.Print("\n")
		if f[0:4] == "pkg/" {
			// we want to install this as a got package!
			installname := fmt.Sprintf("$(pkgdir)/%s", f[4:]+"go")
			dir,_ := path.Split(installname)
			fmt.Printf("%s: %s\n\tmkdir -p %s\n\tcp $< $@\n\n",installname,f+"go",dir)
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
		basename := sourcePath[0:n]
		typesname := sourcePath[n+1:]
		gotname := basename + ".got"
		gotgoname := gotname + "go"
		goname := sourcePath + ".go"
		nip := strings.Index(importPath, "(")
		pkggotname := path.Join(pkgdir,path.Clean(importPath[0:nip]+".gotgo"))

		if fileexists(gotname) {
			fmt.Printf("%s: %s\n\t$<", goname, gotgoname)
		} else if fileexists(pkggotname) {
			fmt.Printf("# looks like we require %s as installed package...\n",
				gotname)
			dir,_ := path.Split(goname)
			fmt.Printf("%s: $(pkgdir)/%s\n\tmkdir -p %s\n\t$<",
				goname, importPath[0:nip]+".gotgo", dir)
		} else {
			fmt.Printf("# I don't know how to make %s from %s or %s\n",
				goname, gotname, pkggotname)
			return
		}
		// We've figured out what gotgo binary to use to create our go
		// file, and now just need the arguments.
		// Begin by scanning the type parameters in the import
		var scan scanner.Scanner
		scan.Init(sourcePath, []byte(typesname), nil, 0)
		types, error := buildit.TypeList(&scan)
		if error != nil { return }
		// Now we want to figure out how to modify the import path The
		// import paths must be written relative to sourcePath, but we
		// only know them relative to the Makefile position.
		relnames := make(map[string]string)
		// ds is the set of directories we might care to strip...
		ds := dirs(path.Clean(goname))
		for k,v := range names {
			gotmismatch := false
			for x:=len(ds)-1;x>=0;x-- {
				if !gotmismatch && v[0:len(ds[x])+1] == ds[x]+"/" {
					v = v[len(ds[x])+1:]
				} else {
					v = "../" + v
				}
			}
			if v[0] != '.' { v = "./" + v }	
			relnames[k] = v
		}
		for _, a := range buildit.GetGotgoArguments(types, relnames) {
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
	switch path.Ext(f) {
	case ".go":
		pname, _, _, _ := getImports(f)
		basename := f[0:len(f)-3]
		if pname == "main" && len(f) > 4 && f[0:4] != "test" {
			mybinfiles += " " + cleanbinname(basename)
		} else if len(f) > 4 && f[0:4] == "pkg/" {
			mypackages += " " + basename + ".a"
		}
	case ".got":
		if len(f) > 4 && f[0:4] == "pkg/" {
			mypackages += " " + f + "go"
		}
	}
}

var mybinfiles = ""
var mypackages = ""
var installbins = ""
var installpkgs = ""
var pkgdir = path.Join(os.Getenv("GOROOT"),"pkg",
	os.Getenv("GOOS")+"_"+os.Getenv("GOARCH"))

func main() {
	path.Walk(".", seeker{}, nil)
	fmt.Printf(`

binaries: %s
packages: %s

include $(GOROOT)/src/Make.$(GOARCH)
ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))

.PHONY: test binaries install installbins installpkgs
.SUFFIXES: .$(O) .go .got .gotgo

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

func taketo(s, sep string) string {
	n := strings.Index(s, sep)
	if n == -1 { return "" }
	return s[0:n]
}

func dirs(s string) []string {
	out := make([]string, 0, len(s))
	s,_ = path.Split(s) // drop filename
	if s == "" { return out }
	s = s[0:len(s)-1]
	for {
		d,ld := path.Split(s)
		out = out[0:len(out)+1]
		out[len(out)-1] = ld
		if d == "" { return out }
		s = d[0:len(d)-1]
	}
	return out
}
