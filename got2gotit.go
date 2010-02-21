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
	typedecs := ""
	vartypes := make(map[string]string)
	for i,t := range types {
		vartypes[params[i]] = t
		typedecs += "type "+ params[i] + " " + t + "\n"
	}
	lastpos := restpos.Offset+1
	e = ioutil.WriteFile(filename+".go",
		strings.Bytes("package "+string(pname)+"\n"+
		string(x[lastpos:])+"\n"+typedecs), 0666)
	if e != nil { return }
	// Now let's write the cool gotit.go file...
	out, e := os.Open(filename+".gotit.go", os.O_WRONLY+os.O_CREAT+os.O_TRUNC,0666)
	if e != nil { return }
	fmt.Fprintf(out, `package main

import ("os"; "fmt"; "unicode"; "strings")

func cleanRune(ch int) int {
	if 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' ||
		ch == '_' || ch >= 0x80 && unicode.IsLetter(ch) ||
		'0' <= ch && ch <= '9' || ch >= 0x80 && unicode.IsDigit(ch) {
		return ch
	}
	return 'ø'
}

func main() {
    pname := "%s"
`, pname)
	for i,t := range types {
		if i > 0 && t == params[0] {
			fmt.Fprintf(out, `
    %s := %s`, params[i], params[0])
		} else {
			fmt.Fprintf(out, `
    %s := "%s"`, params[i], t)
		}
		fmt.Fprintf(out, `
    if len(os.Args) > %d {
        %s = os.Args[%d]
    }
    pname += "龍"+%s
`, i+1, params[i], i+1, params[i]);
	}
	fmt.Fprint(out,`
    fmt.Printf("package %s\n", strings.Map(cleanRune,pname))
`)
	pos,tok,lit := scan.Scan();
	for tok != token.EOF {
		if _,ok := vartypes[string(lit)]; ok {
			chunks := strings.Split(string(x[lastpos:pos.Offset]),"`",0)
			goodchunk := strings.Join(chunks, "`+\"`\"+`")
			fmt.Fprintf(out, "    fmt.Printf(`%s`)\n", goodchunk)
			fmt.Fprintf(out, "    fmt.Print(%s)\n", lit)
			lastpos = pos.Offset + len(lit)
		}
		pos,tok,lit = scan.Scan();
	}
	chunks := strings.Split(string(x[lastpos:]),"`",0)
	goodchunk := strings.Join(chunks, "`+\"`\"+`")
	fmt.Fprintf(out, "    fmt.Printf(`%s`)\n", goodchunk)
	fmt.Fprintf(out, "}\n")
	return
}

func dieOn(e os.Error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) !=  2 {
		dieOn(os.NewError("got2gotit requires one argument"))
	}
	gotname := os.Args[1]
	if gotname[len(gotname)-4:] != ".got" {
		dieOn(os.NewError("the argument should end with .got"))
	}
	dieOn(writeGotGotit(gotname))
	dieOn(buildit.Compile(gotname + ".go"))
	//buildit.CleanCompile(gotname + ".go")
	//os.Remove(gotname+".go")
	dieOn(buildit.Compile(gotname + ".gotit.go"))
	dieOn(buildit.Link(gotname + ".gotit.go"))
	//buildit.CleanCompile(gotname + ".gotit.go")
	//os.Remove(gotname+".gotit.go")
}
