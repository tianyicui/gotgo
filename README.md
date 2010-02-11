gotit
=====

Gotit is two things.  One is a simple runner for .go programs, which
is based on "go-runner", which was created by Dimiter Stanev.  The
other is a prototype for a mechanisms for adding generics to go.  This
page will primarily be about the latter, which is of wider interest,
as there are numerous programs out there for automating the build of
your go files.

Gotit has limited functionality when building programs.  Most notably,
it won't work with locally included packages that are named
differently than their corresponding source files. This implies that
it's one local package per source file.  This limitation should be
largely irrelevant to the generics mechanism.

Quick start
-----------

You can compile `gotit` by just typing

    make

and you can run it on an example program by typing

    ./gotit example.go

which will cause it to compile `example.go` and the two packages
`test.got` and `list.got` and run the resulting executable.

Package templating for generics
-------------------------------

The mechanism for generics that I'm experimenting with it is to
parametrize *packages* according to types.  I am told that this is
similar to the ML module system, but don't know much about that.  It
achieves similar goals to the C++ template approach.

The idea is to create files with a `.got` extention (which stands for
"go template"), which begin with

    package mylist¤type

where the `¤` character in `¤type` makes this not a valid go file (and
thus easy to recognize).  The `.got` file will contain multiple
references to `¤type` in spots where a type is expected.  An ordinary
go file can use this package with an import such as

    import "./mylistøstring"

Note that the `¤type` has been replaced with `østring` in the import.
This defines the type of the list.  When `gotit` encounters this
import, it will make sure the file `mylistøstring.go` has been
created, and build it if necessary.

About the build approach
------------------------

I chose a preprocessing approach, since this seems like the easiest
way to experiment with a small language extension.  Since I had
already been using `go-runner`, it seemed easiest to hard-code the
pre-processing into that tool (which I renamed `gotit`, since it is
now quite changed).

If there is interest in this approach to generics, I expect to create
a stand-alone tool (probably called `got`) that does just the
preprocessing, which could easily be integrated into the standard
Makefile build system (or any other build system).

If this approach is even more successful, and makes it into the
golang.org repository, we'll need to figure out a way to install `got`
files into a standard location, and will need to figure out how and
where to appropriately cache the compiled and preprocessed packages.

About the syntax
----------------

I don't really like my choice of syntax, and if this approach is
adopted as an aspect of go itself, I'd expect a better syntax to be
chosen.  I chose this particular syntax based on a couple of
considerations.  One is that I wanted go files that import a templated
package to be vanilla go files, which meant that the templated import
must be a valid go identifier (letters and numbers).  I arbitrarily
chose the unicode `ø` letter because it'hard to mistake for an ASCII
letter.

In the template itself, I needed a way to specify the template
variables, and wanted something that looks similar to `ø`, but is
*not* a letter or digit, so that the package name would not be a valid
identifier, and the translation should be unambiguous.  I chose `¤`
because it looks similar to `ø`.

I expect neither of these unicode choices to be kept for the long
term.  Ideally we'd have a syntax that *contains* the template types
in some way, perhaps sort of like C++'s syntax.  The entire code for
the `.got` file handling is about thirty lines of code.

Limitations
-----------

- Currently `gotit` can't template using non-builtin types.  To enable
  this, we'll need to have a way to put the '.' in the type name
  (e.g. `os.Error`) so that the package name is a valid identifier.
  e.g. so the import would look something like

      import "./listøosØError"
    
      var errors listøosØError
  
  This will also of course require that the preprocessor add the
  appropriate import statement, which adds a bit more complexity,
  particularly when packages may be imported with alternate names.

- Currently you can only parameterize a package with one type.  It
  won't be hard to extend this to support multiple parameters, but
  this won't mix well with the current simple `ø` prefix, so this will
  require some re-thinking of syntax, ideally by someone with some
  expertise in parsing (esp. in parsing go).

Type checking `got` files
-------------------------

One of my complaints with many (most) generic programming systems
(e.g. cpp macros, C++ templates) is that they generally are not type
checked.  By which I mean that they act as essentially untypechecked
macros, and only the resulting expansion is itself typechecked.

Currently, gotit is subject to the same complaint.  I envision,
however, that it could be extended with relative ease, although I
haven't worked out an appropriate syntax.  But if we came up with a
way to specify that `¤type` must satisfy a given interface (or that it
must be constructable from numeric literals, or must be something that
can be added with `+`), then it should be possible to automatically
create a test file that when compiled verifies that the `got` file
only requires that interface (or those properties).  This would both
enable better (and type checked) documentation of the package, and
verification that additional dependencies don't creep in (e.g. through
use of methods outside that interface).

Problems solved by gotit
------------------------

- In go, there is no way to write a generic data structure, such as a
  linked list.  As the FAQ points out, this isn't *such* a problem,
  since go has a pretty nice stable of built-in data structures, but
  it *is* a problem, since sooner or later you'll run into a need for
  more complicated structures (e.g. perhaps a red-black tree), and you
  don't want to reimplement them for every data type you use.

- In go, there is no way to write a generic function whose return type
  depends on its input.  This is a much more serious shortcoming than
  the previous one, as it means you can't even write generic functions
  for the existing built-in generic data structures.  For example,
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
