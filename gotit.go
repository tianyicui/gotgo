// Copyright 2009 Dimiter Stanev, malkia@gmail.com.
// Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"go/parser"
	"go/scanner"
	"go/ast"
	"strings"
	"strconv"
	"path"
	"./got/buildit"
)

func getmap(m map[string]string, k string) (v string) {
	v, ok := m[k]
	if !ok {
		v = ""
	}
	return
}

var (
	curdir, _ = os.Getwd()
	envbin    = os.Getenv("GOBIN")
	arch      = getmap(map[string]string{"amd64": "6", "386": "8", "arm": "5"}, os.Getenv("GOARCH"))
	compiledAlready = make(map[string]bool)
)

func getLocalImports(filename string) (imports map[string]bool, error os.Error) {
	source, error := ioutil.ReadFile(filename)
	if error != nil {
		return
	}
	file, error := parser.ParseFile(filename, source, nil, parser.ImportsOnly)
	if error != nil {
		return
	}
	for _, importDecl := range file.Decls {
		importDecl, ok := importDecl.(*ast.GenDecl)
		if ok {
			for _, importSpec := range importDecl.Specs {
				importSpec, ok := importSpec.(*ast.ImportSpec)
				if ok {
					for _, importPath := range importSpec.Path {
						importPath, _ := strconv.Unquote(string(importPath.Value))
						if len(importPath) > 0 && importPath[0] == '.' {
							if imports == nil {
								imports = make(map[string]bool)
							}
							dir, _ := path.Split(filename)
							imports[path.Join(dir, path.Clean(importPath))] = true
						}
					}
				}
			}
		}
	}
	return
}

func createGofile(sourcePath string) (error os.Error) {
	switch n := strings.Index(sourcePath, "("); n {
	case -1:
	default:
		basename := sourcePath[0:n]
		typesname := sourcePath[n+1:]
		gotname := basename + ".got"
		gotitname := gotname + "it"
		if needit,_ := shouldUpdate(gotname, gotitname); needit {
			//fmt.Println("Building "+gotitname+" from "+gotname)
			error = buildit.Got2Gotit(gotname)
			if error != nil { return }
		}
		goname := sourcePath + ".go"
		if needit,_ := shouldUpdate(gotitname, goname); needit {
			// first scan the type parameters
			var scan scanner.Scanner
			scan.Init(sourcePath, strings.Bytes(typesname), nil, 0)
			types, error := buildit.TypeList(&scan)
			if error != nil { return }
			error = buildit.GetGofile(sourcePath+".go", basename+".got", types)
			if error != nil { return }
		}
	}
	return
}

func compileRecursively(sourcePath string) (error os.Error) {
	sourcePath = path.Clean(sourcePath)
	if _, exists := compiledAlready[sourcePath]; exists {
		return nil
	}
	localImports, error := getLocalImports(sourcePath+".go")
	if error != nil { return }
	needcompile, _ := shouldUpdate(sourcePath+".go", sourcePath+"."+arch)
	for i, _ := range localImports {
		error = createGofile(i)
		if error != nil { return }
		error = compileRecursively(i)
		if error != nil { return }
		if up, _ := shouldUpdate(i+"."+arch, sourcePath+"."+arch); up {
			needcompile = true
		}
	}
	if needcompile {
		error = buildit.Compile(sourcePath+".go")
	}
	return
}

func shouldUpdate(sourceFile, targetFile string) (doUpdate bool, error os.Error) {
	sourceStat, error := os.Lstat(sourceFile)
	if error != nil {
		return false, error
	}
	targetStat, error := os.Lstat(targetFile)
	if error != nil {
		return true, error
	}
	return targetStat.Mtime_ns < sourceStat.Mtime_ns, error
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "go main-program [arg0 [arg1 ...]]")
		os.Exit(1)
	}

	target := path.Clean(args[0])
	if path.Ext(target) == ".go" {
		target = target[0 : len(target)-3]
	}

	error := compileRecursively(target)
	if error != nil { return } // FIXME: report error

	doLink, _ := shouldUpdate(target+"."+arch, target)
	if doLink {
		buildit.Link(target)
	}
	os.Exec(path.Join(curdir, target), args, os.Environ())
	fmt.Fprintf(os.Stderr, "Error running %v\n", args)
	os.Exit(1)
}
