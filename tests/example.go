package main

// This is a simple example package that uses a template!

import (
	"fmt"
	ints "./demo/slice(int)"
	stringtest "./test(string)"
	inttest "./test(int)"
	list "./demo/list(int)"
	// lists "./demo/slice(list.List)"
)

func main() {
	ss := []string{"hello","world"}
	is := []int{5,4,3,2,1}
	fmt.Println("Head(ss)", stringtest.Head(ss))
	fmt.Println("Tail(is)", inttest.Tail(is))
	// Doesn't the following give you nightmares of lisp?
	fmt.Println(list.Cdr(list.Cons(1,list.Cons(2,list.Cons(3,nil)))))

	ints.Map(func (a int) int { return a*2 }, is)
	fmt.Println("is are now doubled: ", is)
	// ls := []list.List{ list.Car(1,nil), list.Car(2,nil) }
	// fmt.Println("I like lists: ", ls)
}
