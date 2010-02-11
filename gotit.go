// Copyright 2009 Dimiter Stanev, malkia@gmail.com.
// Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"exec"
	"fmt"
	"io/ioutil"
	"go/parser"
	"go/ast"
	"strings"
	"strconv"
	"path"
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

func execp(args []string, dir string) {
	args0, error := exec.LookPath(args[0])
	if error != nil {
		fmt.Fprintf( os.Stderr, "Can't find %s in path: %v\n", args[0], error );
		os.Exit(1);
	}
	p, error := os.ForkExec(args0, args, os.Environ(), dir, []*os.File{os.Stdin, os.Stdout, os.Stderr})
	if error != nil {
		fmt.Fprintf( os.Stderr, "Can't %s\n", error );
		os.Exit(1);
	}
	m, error := os.Wait(p, 0)
	if error != nil {
		fmt.Fprintf( os.Stderr, "Can't %s\n", error );
		os.Exit(1);
	}
	if m.ExitStatus() != 0 {
		os.Exit(int(m.ExitStatus()));
	}
}

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
	switch n := strings.Index(sourcePath, "ø"); n {
	case -1:
	default:
		basename := sourcePath[0:n]
		typename := sourcePath[n+2:]
		gotname := basename + ".got"
		goname := sourcePath + ".go"
		if needit,_ := shouldUpdate(gotname, goname); needit {
			templbytes, error := ioutil.ReadFile(gotname)
			templ := string(templbytes)
			if error != nil { return }
			packagewords := "package "+basename+"¤"
			n := strings.Index(templ, packagewords)
			if n == -1 {
				fmt.Println("Bad template file", gotname)
				return os.NewError("Bad template file "+gotname)
			}
			nl := strings.Index(templ[n+len(packagewords):], "\n")
			key := "¤"+templ[n+len(packagewords):n+len(packagewords)+nl]

			// Now we want to place every instance of key with typename
			// except for the first...
			bits := strings.Split(templ,key,0)
			out := bits[0] + "ø"
			for _,v := range bits[1:] {
				out += typename + v
			}
			error = ioutil.WriteFile(goname, strings.Bytes(out), 0644)
		}
	}
	return
}

func compileRecursively(sourcePath string) (error os.Error) {
	// We should consider ¤ and ø from ctrl-x 8 $ and ctrl-x 8 /o respectively
	sourcePath = path.Clean(sourcePath)
	if _, exists := compiledAlready[sourcePath]; exists {
		return nil
	}
	localImports, error := getLocalImports(sourcePath+".go")
	if error != nil { return }
	needcompile, _ := shouldUpdate(sourcePath+".go", sourcePath+"."+arch)
	for i, _ := range localImports {
		createGofile(i)
		compileRecursively(i)
		if up, _ := shouldUpdate(i+"."+arch, sourcePath+"."+arch); up {
			needcompile = true
		}
	}
	if needcompile {
		dir, filename := path.Split(sourcePath)
		execp([]string{path.Join(envbin, arch+"g"), filename + ".go"}, dir)
	}
	return nil
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

	compileRecursively(target)

	doLink, _ := shouldUpdate(target+"."+arch, target)
	if doLink {
		execp([]string{path.Join(envbin, arch+"l"), "-o", target, target+"."+arch}, "")
	}
	os.Exec(path.Join(curdir, target), args, os.Environ())
	fmt.Fprintf(os.Stderr, "Error running %v\n", args)
	os.Exit(1)
}
