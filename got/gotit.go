package gotit

import (
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"go/token"
	"go/scanner"
	"./buildit"
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
	out, e := os.Open(filename+"it.go", os.O_WRONLY+os.O_CREAT+os.O_TRUNC,0666)
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
    imports := ""
    firstarg := 1
    for firstarg=1; firstarg<len(os.Args); firstarg++ {
        if os.Args[firstarg] == "--import" {
            firstarg++
            imports += "\n"+os.Args[firstarg]
        } else if len(os.Args[firstarg]) > len("--import=") &&
                  os.Args[firstarg][0:len("--import=")] == "--import=" {
            imports += "\n"+os.Args[firstarg][len("--import="):]
        } else {
            break;
        }
    }
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
    if len(os.Args) > firstarg + %d - 1 {
        %s = os.Args[firstarg + %d - 1]
    }
    pname += "龍"+%s
`, i+1, params[i], i+1, params[i]);
	}
	fmt.Fprint(out,`
    fmt.Printf("package %s\n", strings.Map(cleanRune,pname))
    fmt.Printf("%s\n", imports)
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

	fmt.Fprint(out,`
    fmt.Print("// Here we will test that the types parameters are ok...\n")
    fmt.Printf("\n\nfunc testTypes(arg0 %s`)
	for i,_ := range types[1:] { fmt.Fprintf(out, `, arg%d %%s`, i+1) }
	fmt.Fprint(out, `) {\n"`)
	for _,p := range params { fmt.Fprintf(out, `, %s`, p) }
	fmt.Fprint(out, `)`)
	fmt.Fprintf(out, `
    fmt.Print("    f := func(%s`, types[0])
	for _,t := range types[1:] {
		if t == params[0] { t = types[0] }
		fmt.Fprintf(out, `, %s`, t)
	}
  fmt.Fprint(out, `) { } // this func does nothing...\n")`)
	convert := func(t string, argnum int) string {
		arg := "arg" + fmt.Sprint(argnum)
		if strings.Index(t, "{") == -1 {
			return t+"("+arg+")"
		}
		return arg  // it's an interface, so we needn't convert
	}
	fmt.Fprintf(out, `
    fmt.Print("    f(%s`, convert(types[0], 0))
	for i,t := range types[1:] {
		if t == params[0] { t = types[0] }
		fmt.Fprintf(out, `, %s`, convert(t, i+1))
	}
	fmt.Fprint(out, `")
    fmt.Print(")\n}\n")
`)

	fmt.Fprint(out, "    // End of main function!\n")
	fmt.Fprint(out, "}\n")
	return
}

func dieOn(e os.Error) {
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
		os.Exit(1)
	}
}

func Gotit(gotname string) (error os.Error) {
	if gotname[len(gotname)-4:] != ".got" {
		return os.NewError("the argument should end with .got")
	}
	error = writeGotGotit(gotname)
	if error != nil { return }
	error = buildit.Compile(gotname + ".go")
	if error != nil { return }
	//buildit.CleanCompile(gotname + ".go")
	//os.Remove(gotname+".go")
	error = buildit.Compile(gotname + "it.go")
	if error != nil { return }
	error = buildit.Link(gotname + "it.go")
	//buildit.CleanCompile(gotname + "it.go")
	//os.Remove(gotname+"it.go")
	return
}
