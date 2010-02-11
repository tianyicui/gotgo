package main

// This is a simple example package that uses a template!

import (
	"fmt"
	"./testøstring"
	"./testøint"
	list "./listøint"
)

func main() {
	ss := []string{"hello","world"}
	is := []int{5,4,3,2,1}
	fmt.Println("Head(ss)", testøstring.Head(ss))
	fmt.Println("Tail(is)", testøint.Tail(is))
	// Doesn't the following give you nightmares of lisp?
	fmt.Println(list.Cdr(list.Cons(1,list.Cons(2,list.Cons(3,nil)))))
}
