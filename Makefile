# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.


all: Makefile  bin/gotgo bin/gotimports bin/mkmake tests/demo/list.gotgo tests/demo/slice.gotgo tests/example tests/test.gotgo

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: test
.SUFFIXES: .$(O) .go .got .gotgo

Makefile: bin/mkmake
	./bin/mkmake > $@
.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotgo:
	gotgo "$<"

src/got/buildit.$(O): src/got/buildit.go

# src/got/gotgo.go imports src/got/buildit
src/got/gotgo.$(O): src/got/gotgo.go src/got/buildit.$(O)

# src/gotgo.go imports src/got/gotgo
bin/gotgo: src/gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
src/gotgo.$(O): src/gotgo.go src/got/gotgo.$(O)

# src/gotimports.go imports src/got/gotgo
# src/gotimports.go imports src/got/buildit
bin/gotimports: src/gotimports.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
src/gotimports.$(O): src/gotimports.go src/got/buildit.$(O) src/got/gotgo.$(O)

# src/mkmake.go imports src/got/buildit
bin/mkmake: src/mkmake.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
src/mkmake.$(O): src/mkmake.go src/got/buildit.$(O)

# found file tests/demo/finalizer.got to build...
# error:  tests/demo/finalizer.got:6:18: expected ';', found '('
tests/demo/finalizer.gotgo: tests/demo/finalizer.got 

tests/demo/list(int).$(O): tests/demo/list(int).go

# found file tests/demo/list.got to build...
# error:  tests/demo/list.got:1:13: expected ';', found '('
tests/demo/list.gotgo: tests/demo/list.got 

tests/demo/list.got.$(O): tests/demo/list.got.go

tests/demo/list.gotgo: tests/demo/list.gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/demo/list.gotgo.$(O): tests/demo/list.gotgo.go

tests/demo/slice(int).$(O): tests/demo/slice(int).go

# tests/demo/slice(list.List).go imports tests/demo/list(int)
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
tests/demo/slice(list.List).$(O): tests/demo/slice(list.List).go tests/demo/list(int).$(O)

# found file tests/demo/slice.got to build...
# error:  tests/demo/slice.got:1:14: expected ';', found '('
tests/demo/slice.gotgo: tests/demo/slice.got 

tests/demo/slice.got.$(O): tests/demo/slice.got.go

tests/demo/slice.gotgo: tests/demo/slice.gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/demo/slice.gotgo.$(O): tests/demo/slice.gotgo.go

# tests/example.go imports tests/demo/slice(list.List)
tests/demo/slice(list.List).go: tests/demo/slice.gotgo
	$< '--import' 'import list "../../tests/demo/list(int)"' 'list.List' > "$@"
# tests/example.go imports tests/demo/list(int)
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
# tests/example.go imports tests/test(string)
tests/test(string).go: tests/test.gotgo
	$< 'string' > "$@"
# tests/example.go imports tests/demo/slice(int)
tests/demo/slice(int).go: tests/demo/slice.gotgo
	$< 'int' > "$@"
# tests/example.go imports tests/test(int)
tests/test(int).go: tests/test.gotgo
	$< 'int' > "$@"
tests/example: tests/example.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/example.$(O): tests/example.go tests/demo/list(int).$(O) tests/demo/slice(int).$(O) tests/demo/slice(list.List).$(O) tests/test(int).$(O) tests/test(string).$(O)

tests/test(int).$(O): tests/test(int).go

tests/test(string).$(O): tests/test(string).go

# found file tests/test.got to build...
# error:  tests/test.got:1:13: expected ';', found '('
tests/test.gotgo: tests/test.got 

tests/test.got.$(O): tests/test.got.go

tests/test.gotgo: tests/test.gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/test.gotgo.$(O): tests/test.gotgo.go

