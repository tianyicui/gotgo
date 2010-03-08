# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all: Makefile binaries

Makefile: bin/gotmake scripts/Make.header
	cp -f scripts/Make.header $@
	./bin/gotmake >> $@

test: tests/example

install: installbins installpkgs


binaries:  pkg/gotgo/slice.gotgo bin/gotgo bin/gotimports bin/gotmake
packages:  pkg/gotgo/slice.gotgo pkg/gotgo/slice.got.a

include $(GOROOT)/src/Make.$(GOARCH)
ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))

.PHONY: test binaries install installbins installpkgs
.SUFFIXES: .$(O) .go .got .gotgo

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotgo:
	gotgo "$<"

# found file pkg/gotgo/slice.got to build...
# error:  pkg/gotgo/slice.got:1:14: expected ';', found '('
pkg/gotgo/slice.gotgo: pkg/gotgo/slice.got 
$(pkgdir)/gotgo/slice.gotgo: pkg/gotgo/slice.gotgo
	mkdir -p $(pkgdir)/gotgo/
	cp $< $@


pkg/gotgo/slice.got.$(O): pkg/gotgo/slice.got.go
$(pkgdir)/pkg/gotgo/slice.got.$(O): pkg/gotgo/slice.got.$(O)
	mkdir -p $(pkgdir)/pkg/gotgo/
	cp $< $@


# ignoring pkg/gotgo/slice.gotgo.go, since it's a generated file
src/got/buildit.$(O): src/got/buildit.go

# src/got/gotgo.go imports src/got/buildit
src/got/gotgo.$(O): src/got/gotgo.go src/got/buildit.$(O)

# src/gotgo.go imports src/got/gotgo
bin/gotgo: src/gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotgo: bin/gotgo
	cp $< $@
src/gotgo.$(O): src/gotgo.go src/got/gotgo.$(O)

# src/gotimports.go imports src/got/gotgo
# src/gotimports.go imports src/got/buildit
bin/gotimports: src/gotimports.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotimports: bin/gotimports
	cp $< $@
src/gotimports.$(O): src/gotimports.go src/got/buildit.$(O) src/got/gotgo.$(O)

# src/gotmake.go imports src/got/buildit
bin/gotmake: src/gotmake.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotmake: bin/gotmake
	cp $< $@
src/gotmake.$(O): src/gotmake.go src/got/buildit.$(O)

# found file tests/demo/finalizer.got to build...
# error:  tests/demo/finalizer.got:6:18: expected ';', found '('
tests/demo/finalizer.gotgo: tests/demo/finalizer.got 

tests/demo/list(int).$(O): tests/demo/list(int).go

# found file tests/demo/list.got to build...
# error:  tests/demo/list.got:1:13: expected ';', found '('
tests/demo/list.gotgo: tests/demo/list.got 

tests/demo/list.got.$(O): tests/demo/list.got.go

# ignoring tests/demo/list.gotgo.go, since it's a generated file
tests/demo/slice(int).$(O): tests/demo/slice(int).go

# tests/demo/slice(list.List).go imports tests/demo/list(int)
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
tests/demo/slice(list.List).$(O): tests/demo/slice(list.List).go tests/demo/list(int).$(O)

# found file tests/demo/slice.got to build...
# error:  tests/demo/slice.got:1:14: expected ';', found '('
tests/demo/slice.gotgo: tests/demo/slice.got 

tests/demo/slice.got.$(O): tests/demo/slice.got.go

# ignoring tests/demo/slice.gotgo.go, since it's a generated file
# tests/example.go imports tests/demo/list(int)
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
# tests/example.go imports tests/test(string)
tests/test(string).go: tests/test.gotgo
	$< 'string' > "$@"
# tests/example.go imports tests/gotgo/slice(int)
# looks like we require tests/gotgo/slice.got as installed package...
tests/gotgo/slice(int).go: $(pkgdir)/./gotgo/slice.gotgo
	mkdir -p tests/gotgo/
	$< 'int' > "$@"
# tests/example.go imports tests/test(int)
tests/test(int).go: tests/test.gotgo
	$< 'int' > "$@"
# tests/example.go imports tests/gotgo/slice(list.List)
# looks like we require tests/gotgo/slice.got as installed package...
tests/gotgo/slice(list.List).go: $(pkgdir)/./gotgo/slice.gotgo
	mkdir -p tests/gotgo/
	$< '--import' 'import list "../demo/list(int)"' 'list.List' > "$@"
tests/example: tests/example.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/example.$(O): tests/example.go tests/demo/list(int).$(O) tests/gotgo/slice(int).$(O) tests/gotgo/slice(list.List).$(O) tests/test(int).$(O) tests/test(string).$(O)

tests/gotgo/slice(int).$(O): tests/gotgo/slice(int).go

# tests/gotgo/slice(list.List).go imports tests/demo/list(int)
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
tests/gotgo/slice(list.List).$(O): tests/gotgo/slice(list.List).go tests/demo/list(int).$(O)

tests/test(int).$(O): tests/test(int).go

tests/test(string).$(O): tests/test(string).go

# found file tests/test.got to build...
# error:  tests/test.got:1:13: expected ';', found '('
tests/test.gotgo: tests/test.got 

tests/test.got.$(O): tests/test.got.go

# ignoring tests/test.gotgo.go, since it's a generated file
installbins:  $(bindir)/gotgo $(bindir)/gotimports $(bindir)/gotmake
installpkgs:  $(pkgdir)/gotgo/slice.gotgo $(pkgdir)/pkg/gotgo/slice.got.$(O)
