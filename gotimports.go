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
	"./got/gotgo"
)

var (
	curdir, _ = os.Getwd()
)

func getImports(filename string) (imports map[string]bool, names map[string]string, error os.Error) {
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
					importPath, _ := strconv.Unquote(string(importSpec.Path.Value))
					if len(importPath) > 0 {
						if imports == nil {
							imports = make(map[string]bool)
							names = make(map[string]string)
						}
						dir, _ := path.Split(filename)
						imports[path.Join(dir, path.Clean(importPath))] = true
						names[importSpec.Name.String()] =
							path.Join(dir, path.Clean(importPath))
					}
				}
			}
		}
	}
	return
}

func createGofile(sourcePath string, importMap map[string]string) (error os.Error) {
	switch n := strings.Index(sourcePath, "("); n {
	case -1:
		// ignore non-templated import...
	default:
		basename := sourcePath[0:n]
		typesname := sourcePath[n+1:]
		gotname := basename + ".got"
		gotgoname := gotname + "go"
		if needit,_ := shouldUpdate(gotname, gotgoname); needit {
			error = gotgo.Gotgo(gotname)
			if error != nil { return explain("gotgo.Gotgo", error) }
		}
		goname := sourcePath + ".go"
		if needit,_ := shouldUpdate(gotgoname, goname); needit {
			// first scan the type parameters
			var scan scanner.Scanner
			scan.Init(sourcePath, strings.Bytes(typesname), nil, 0)
			types, error := buildit.TypeList(&scan)
			if error != nil { return }
			error = buildit.GetGofile(sourcePath+".go", basename+".got", types,
				importMap)
			if error != nil { return explain("GetGofile", error) }
		}
	}
	return
}

func explain(x string, e os.Error) os.Error {
	return os.NewError(x + ": " + e.String())
}

func createImports(sourcePath string) (error os.Error) {
	sourcePath = path.Clean(sourcePath)
	allImports, importMap, error := getImports(sourcePath+".go")
	if error != nil { return explain("getImports "+sourcePath+".go\n", error) }
	for i, _ := range allImports {
		error = createGofile(i, importMap)
		if error != nil { return explain("createGofile", error) }
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
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "go main-program GOFILE")
		os.Exit(1)
	}
	
	target := path.Clean(os.Args[1])
	if path.Ext(target) == ".go" {
		target = target[0 : len(target)-3]
	}

	error := createImports(target)
	if error != nil {
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}
}
