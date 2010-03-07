package main

import (
	"os"
	"fmt"
	"./got/gotgo"
)

func main() {
	if len(os.Args) !=  2 {
		fmt.Fprintln(os.Stderr, "gotit requires one argument")
		os.Exit(1)
	}
	error := gotgo.Gotgo(os.Args[1])
	if error != nil {
		fmt.Fprintln(os.Stderr, error)
		os.Exit(1)
	}
}
