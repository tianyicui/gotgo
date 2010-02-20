package main

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"go/token"
	"go/scanner"
	"./got/buildit"
)

func writeGotGotit(filename string) (e os.Error) {
	x, e := ioutil.ReadFile(filename)
	if e != nil { return }
	var scan scanner.Scanner
	scan.Init(filename, x, nil, 0)
	tok := token.COMMA // anything but EOF or PACKAGE
	for tok != token.EOF && tok != token.PACKAGE {
		_, tok, _ = scan.Scan()
	}
	if tok == token.EOF { return os.NewError("Unexpected EOF...") }
	_, tok, pname := scan.Scan()
	if tok != token.IDENT {
		return os.NewError("Expected package ident, not "+string(pname))
	}
	_, tok, lit := scan.Scan()
	if tok != token.LPAREN {
		return os.NewError("Expected (, not "+string(lit))
	}
	params, types, restpos, e := buildit.GetTypes(&scan)
	if e != nil { return }
	fmt.Println("See: ",params, " and ",types)
	typedecs := ""
	for i,t := range types {
		typedecs += "type "+ params[i] + " " + t + "\n"
	}
	e = ioutil.WriteFile(filename+".go",
		strings.Bytes("package "+string(pname)+"\n"+
		string(x[restpos.Offset+1:])+"\n"+typedecs), 0666)
	if e != nil { return }
	// Now let's write the cool gotit.go file...
	
	return
}

func dieOn(e os.Error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
}

func main() {
	gotname := os.Args[1]
	dieOn(writeGotGotit(gotname))
	buildit.GetPackageName("foo", nil)
}
