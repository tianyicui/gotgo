Gotit, Mark II
==============

This document describes the second iteration of my attempt at a
reasoonable implementation of generics for [go](http://golang.org)
based on the idea of template packages.

Quick start
-----------

You can compile `gotit` by just typing

    make

and you can run it on an example program by typing

    ./go tests/example.go

or

    make tests/example
    ./tests/example

either of which will compile `example.go` and the three packages
`tests/test.got`, `tests/demo/slice.got` and `tests/demo/list.got` and
run the resulting executable.  You can examine the [Makefile][1] to
see how to integrate `gotit` into your own build system.

If you just want to see what a template package will look like, check
out [slice.got][2], which is a simple package exporting handy
functions for slices, such as Map, Fold, Filter, Append, Cat (concat).

[1]: http://github.com/droundy/gotit/blob/master/Makefile
[2]: http://github.com/droundy/gotit/blob/master/tests/demo/slice.got

The got file
============

A template package is contained in a got file, ending with the `.got`
extension.  As you might be able to guess, "got" stands for "go
template".  In the future, this should be extended so a template
package can have multiple source files, just like ordinary packages.

The got file differs from an ordinary go file only in its package
line, which will be something like

    package slice(type a, type b a)

This defines a package template named "slice", with two type
parameters.  The first could be any type, which will be called "a",
and the second type, called "b" will default to be the same as "a", if
no second type parameter is provided.  The default type argument is
optional, and will itself default to interface{}.  The default type
should be either an interface type, or a type defined in terms of
other template parameters, although I'm not sure whether/how to
enforce this.  Your got file *must* be able to compile with the
default type!

Within this package, the types "a" and "b" will be in scope, so the
slice got file might contain:

    func Append(slice []a, datum a) []a {
        // first expand if needed...
        slice = slice[0:len(slice)+1]
        slice[len(slice)-1] = datum
        return slice
    }

We could also have a function that uses both "a" and "b":

    func Map(f func(a) b, slice []a) []b {
        out := make([]b, len(slice))
        for i,v := range slice {
            out[i] = f(v)
        }
        return out
    }

Importing a got package
=======================

If you want to use the above package, you will do so by creating an
ordinary go file that has a special import statement that specifies
the types desired.  So, given the above got file, we could write

    import intslice "./slice(int)"

    func foo(x int, xs []int, ...) {
        ...
        intslice.Append(xs,x)

This import is valid go (since the language spec states that the
interpretation of the import string is implementation-dependent), but
I'm not sure if it will be accepted by the go compilers.  I'm hoping
that since it's a valid path name on most file systems, that it will
work just fine.

The package name of this templated import will be dependent on the
gotit implementation, so you need to specify your own qualtified
package name in order to safely use the package.

Getting fancier with interface types
====================================

We can get a bit fancier in the package specification by using
interfaces.

    package list(type a fmt.Stringer)

    import fmt "fmt"

    type T struct {
        Datum a
        Next *T
    }
    func (l T) String() String {
        if l.Next != nil {
            return l.Datum.String() + ", " + l.Next.String()
        } else {
            return l.Datum.String()
        }
    }

This isn't much different from what I've already described, but there
is (will be) an additional twist that isn't shown, which is that if
you make the default type an interface, then any type parameter passed
to the template *must satisfy that interface*! This is valuable
because it means that you can guarantee that you can't break the
compile of any client code unless you change the template signature.

Still on the TODO list is figuring out how to properly handle
templates that require a builtin type, such as

    package maxmin(type a)
    
    func max(x,y a) a {
        if x > y { return x }
        return y
    }

This won't work on struct types, and we ought to be able to specify
that in the package signature.

How does it work?
=================

I have implemented a simple got compiler (called gotit, barring better
ideas), which you pass the name of the got file you wish to compile.
Gotit will parse this file, and first generate a go file for the
default types, which it will compile---to make sure it compiles, and
so you can get easy and immediate feedback if you break something.
This is intended to make the "got" templating language itself closer
to statically typed than cpp macros or C++ templates.

After compiling the template with its default types, a go program will
be written (by gotit) which accepts as its arguments a list of
parameter types and eventually flags indicating any additional imports
needed to define those types, which will first verify that its
arguments satisfy the interface constraints in the package, and then
will write to stdout a go file that can be compiled as that template
package.

A go build system needs to track down and build imported packages, and
in order to work with gotit, it needs to know how to build templated
packages.  An example of such a build system is provided in the go.go
program, which both builds and runs simple (possibly templated) go
programs.


Limitations
-----------

- Currently `gotit` can't template using non-builtin types, because it
  doesn't know how to add the necessary imports.  To enable
  this, I'll need to add a flag to `gotit` to specify imports, and the
  build system will have to figure out the required imports from the
  import statement, e.g. from

      import foolist "./demo/list(data.Foo)"

  it would need to figure out what the `data` import means.

- Type-checking `got` files.  I haven't  yet implemented a check that
  the provided type satisfies the requested interface.  I plan to do
  this by adding to the generated go file a small unused function that
  does something like

      func unusedForGotIt(data a) {
          f := func(defaultTypeForA) { }
          f(defaultTypeForA(a))
      }

  where `defaultTypeForA` is the "default type" specified for the type
  parameter `a`.  This will verify that the given type is indeed
  [assignment-compatible][3]
  with the default type.  For an interface as the default type, this
  will mean that the requested type satisfies the interface, and for
  primitives (e.g. int) as the default type, it means that the
  specified type is "similar" in some way that I only vaguely
  understand (but hopefully is helpful).

[3]: http://golang.org/doc/go_spec.html#Assignment_compatibility

Problems solved by gotit
------------------------

- In go, there is no way to write a generic data structure, such as a
  linked list.  As the FAQ points out, this isn't *such* a problem,
  since go has a pretty nice stable of built-in data structures, but
  it *is* a problem, since sooner or later you'll run into a need for
  more complicated structures (e.g. perhaps a red-black tree), and you
  don't want to reimplement them for every data type you use.

- In go, there is no way to write a generic function whose return type
  depends on its input, or with two arguments of the same (but
  variable type).  This is a much more serious shortcoming than the
  previous one, as it means you can't even write generic functions for
  the existing built-in generic data structures.  For example,
  consider the generic version of

      func Append(x int, xs []int) []int

  which will append the given value to the slice, expanding it if
  needed.  This is precisely the sort of function you'd like to write
  only once (or twice perhaps, to treat "large" data types
  differently), but currently copy-and-paste is the only way to reuse
  this code.  (And similarly for methods...)

Problems not solved by gotit
----------------------------

- Overloading.  One might like to have a single function that operates
  on a variety of types.  Gotit doesn't give you this.  You only have
  to write one version of the function to handle multiple types, but
  when you *use* the function, you always have to be explicit about
  which types it's operating on.
