Gotit, Mark II
==============

This document describes the second iteration of my attempt at a
reasoonable implementation of generics for [go](http://golang.org)
based on the idea of template packages.

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
is an additional twist that isn't shown, which is that if you make the
default type an interface, then any type parameter passed to the
template *must satisfy that interface*! This is valuable because it
means that you can guarantee that you can't break the compile of any
client code unless you change the template signature.

Still on the TODO list is figuring out how to handle templates that
require a builtin type, such as

    package maxmin(type a)
    
    func max(x,y a) a {
        if x > y { return x }
        return y
    }

This won't work on non-primitive types, and we ought to be able to
specify that in the package signature.

How does it work?
=================

It doesn't.  I haven't yet implemented this version.

How will it work?
=================

I will implement a got compiler (probably called gotit, barring better
ideas), which you pass the name of the gotit file you wish to compile.
Gotit will parse this file, and first generate a go file for the
default types, which it will compile---to make sure it compiles, and
so you can get easy and immediate feedback if you break something.
This is intended to make the "got" templating language itself
statically typed, unlike cpp macros or C++ templates.

After compiling the template with its default types, a go program will
be written (by gotit) which accepts as its arguments a list of
parameter types and probably flags indicating any additional imports
needed too define those types, which will first verify that its
arguments satisfy the interface constraints in the package, and then
will write to stdout a go file that can be compiled as that template
package.

A go build system needs to track down and build imported packages, and
in order to work with gotit, it needs to know how to build templated
packages.  I won't write more on this until I've gotten into the
actual implementation...
