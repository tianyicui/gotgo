// Copyright 2009 Dimiter Stanev, malkia@gmail.com.
// Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package buildit

import (
	"os"
	"fmt"
	"exec"
	"path"
	"strings"
	"bytes"
	"go/scanner"
	"go/token"
	"go/ast"
	"go/parser"
	"go/printer"
	stringslice "../gotgo/slice(string)"
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

func Explain(x string, e os.Error) os.Error {
	return os.NewError(x + ": " + e.String())
}

func execp(dir string, args ...string) (error os.Error) {
	args0, error := exec.LookPath(args[0])
	if error != nil { return Explain(args[0], error) }
	p, error := os.ForkExec(args0, args, os.Environ(), dir, []*os.File{os.Stdin, os.Stdout, os.Stderr})
	if error != nil { return Explain("os.ForkExec",error) }
	m, error := os.Wait(p, 0)
	if error != nil { return }
	if m.ExitStatus() != 0 {
		return os.NewError("Command failed with: "+string(int(m.ExitStatus())));
	}
	return
}

func execpout(out, dir string, args []string) (error os.Error) {
	args0, error := exec.LookPath(args[0])
	if error != nil {
		if _,e := os.Lstat(args[0]); e != nil {
			return Explain("Couldn't find in path "+args[0], error)
		}
		args0 = args[0] // the file is locally available
	}
	outf, error := os.Open(out, os.O_WRONLY+os.O_CREAT+os.O_TRUNC, 0666)
	if error != nil { return Explain("Trouble creating "+out, error) }
	p, error := os.ForkExec(args0, args, os.Environ(), dir, []*os.File{os.Stdin, outf, os.Stderr})
	if error != nil { return Explain("os.ForkExec", error) }
	m, error := os.Wait(p, 0)
	if error != nil { return Explain("os.Wait", error) }
	if m.ExitStatus() != 0 {
		return os.NewError("Command failed with: "+string(int(m.ExitStatus())));
	}
	return
}

func Compile(sourcePath string) (error os.Error) {
	if len(sourcePath) > 3 && sourcePath[len(sourcePath)-3:] == ".go" {
		sourcePath = sourcePath[0:len(sourcePath)-3]
	}
	sourcePath = path.Clean(sourcePath)
	dir, filename := path.Split(sourcePath)
	return execp(dir, path.Join(envbin, arch+"g"), filename+".go")
}

func CleanCompile(sourcePath string) (error os.Error) {
	if len(sourcePath) > 3 && sourcePath[len(sourcePath)-3:] == ".go" {
		sourcePath = sourcePath[0:len(sourcePath)-3]
	}
	sourcePath = path.Clean(sourcePath)
	return os.Remove(sourcePath+"."+arch)
}

func Link(sourcePath string) (error os.Error) {
	if len(sourcePath) > 3 && sourcePath[len(sourcePath)-3:] == ".go" {
		sourcePath = sourcePath[0:len(sourcePath)-3]
	}
	sourcePath = path.Clean(sourcePath)
	dir, filename := path.Split(sourcePath)
	return execp(dir, path.Join(envbin, arch+"l"), "-o", filename, filename+"."+arch)
}

func getType(s *scanner.Scanner) (t string, pos token.Position, tok token.Token, lit []byte)  {
	pos, tok, lit = s.Scan()
	for tok != token.RPAREN && tok != token.COMMA {
		t += string(lit)
		pos, tok, lit = s.Scan()
	}
	if t == "" { t = "interface{}" }
	return
}

func GetTypes(s *scanner.Scanner) (params []string, types []string, pos token.Position, error os.Error) {
	tok := token.COMMA
	var lit []byte
	for tok == token.COMMA {
		pos, tok,lit = s.Scan()
		if tok != token.TYPE {
			error = os.NewError("Expected 'type', not "+string(lit))
			return
		}
		var tname string
		var par []byte
		pos, tok,par = s.Scan()
		if tok != token.IDENT {
			error = os.NewError("Identifier expected, not "+string(par))
			return
		}
		tname,pos,tok,lit = getType(s)
		params = stringslice.Append(params, string(par))
		types = stringslice.Append(types, string(tname))
	}
	if tok != token.RPAREN {
		error = os.NewError(fmt.Sprintf("inappropriate token %v with lit: %s",
			tok, lit))
	}
	return
}

func TypeList(s *scanner.Scanner) (types []string, error os.Error) {
	tok := token.COMMA
	var lit []byte
	for tok == token.COMMA {
		var tname string
		tname,_,tok,lit = getType(s)
		types = stringslice.Append(types, string(tname))
	}
	if tok != token.RPAREN {
		error = os.NewError(fmt.Sprintf("inappropriate token %v with lit: %s",
			tok, lit))
	}
	return
}

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
