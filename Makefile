# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.

all: Makefile binaries

Makefile: bin/gotmake scripts/make.header $(wildcard *.go) $(wildcard *.got)
	cp -f scripts/make.header $@
	./bin/gotmake >> $@

test: tests/example

install: installbins installpkgs


include $(GOROOT)/src/Make.$(GOARCH)

binaries:  pkg/gotgo/finalizer.gotgo pkg/gotgo/slice.gotgo bin/gotgo bin/gotmake
packages:  pkg/gotgo/finalizer.gotgo pkg/gotgo/slice.gotgo

ifndef GOBIN
GOBIN=$(HOME)/bin
endif

# ugly hack to deal with whitespaces in $GOBIN
nullstring :=
space := $(nullstring) # a space at the end
bindir=$(subst $(space),\ ,$(GOBIN))
pkgdir=$(subst $(space),\ ,$(GOROOT)/pkg/$(GOOS)_$(GOARCH))

.PHONY: test binaries packages install installbins installpkgs
.SUFFIXES: .$(O) .go .got .gotgo

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotgo:
	gotgo "$<"

# found file pkg/gotgo/finalizer.got to build...
# error:  pkg/gotgo/finalizer.got:6:18: expected ';', found '('
pkg/gotgo/finalizer.gotgo: pkg/gotgo/finalizer.got
$(pkgdir)/gotgo/finalizer.gotgo: pkg/gotgo/finalizer.gotgo
	mkdir -p $(pkgdir)/gotgo/
	cp $< $@


# ignoring pkg/gotgo/finalizer.got.go, since it's a generated file
# ignoring pkg/gotgo/finalizer.gotgo.go, since it's a generated file
# found file pkg/gotgo/slice.got to build...
# error:  pkg/gotgo/slice.got:1:14: expected ';', found '('
pkg/gotgo/slice.gotgo: pkg/gotgo/slice.got
$(pkgdir)/gotgo/slice.gotgo: pkg/gotgo/slice.gotgo
	mkdir -p $(pkgdir)/gotgo/
	cp $< $@


# ignoring pkg/gotgo/slice.got.go, since it's a generated file
# ignoring pkg/gotgo/slice.gotgo.go, since it's a generated file
# I don't know how to make src/gotgo/slice(string).go from src/gotgo/slice.got or /mnt/home/droundy/src/go/pkg/gotgo/slice.gotgo
src/got/buildit.$(O): src/got/buildit.go src/gotgo/slice(string).$(O)

src/got/gotgo.$(O): src/got/gotgo.go src/got/buildit.$(O)

src/gotgo/slice(string).$(O): src/gotgo/slice(string).go

bin/gotgo: src/gotgo.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotgo: bin/gotgo
	cp $< $@
src/gotgo.$(O): src/gotgo.go src/got/gotgo.$(O)

# looks like we require src/gotgo/slice.got as installed package...
src/gotgo/slice(string).go: $(pkgdir)/./gotgo/slice.gotgo
	mkdir -p src/gotgo/
	$< 'string' > "$@"
bin/gotmake: src/gotmake.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
$(bindir)/gotmake: bin/gotmake
	cp $< $@
src/gotmake.$(O): src/gotmake.go src/got/buildit.$(O) src/gotgo/slice(string).$(O)

tests/demo/list(int).$(O): tests/demo/list(int).go

# found file tests/demo/list.got to build...
# error:  tests/demo/list.got:1:13: expected ';', found '('
tests/demo/list.gotgo: tests/demo/list.got

# ignoring tests/demo/list.got.go, since it's a generated file
# ignoring tests/demo/list.gotgo.go, since it's a generated file
tests/demo/list(int).go: tests/demo/list.gotgo
	$< 'int' > "$@"
tests/test(string).go: tests/test.gotgo
	$< 'string' > "$@"
# looks like we require tests/gotgo/slice.got as installed package...
tests/gotgo/slice(int).go: $(pkgdir)/./gotgo/slice.gotgo
	mkdir -p tests/gotgo/
	$< 'int' > "$@"
tests/test(int).go: tests/test.gotgo
	$< 'int' > "$@"
# looks like we require tests/gotgo/slice.got as installed package...
tests/gotgo/slice(list.List).go: $(pkgdir)/./gotgo/slice.gotgo
	mkdir -p tests/gotgo/
	$< --import 'import list "../demo/list(int)"' 'list.List' > "$@"
tests/example: tests/example.$(O)
	@mkdir -p bin
	$(LD) -o $@ $<
tests/example.$(O): tests/example.go tests/demo/list(int).$(O) tests/gotgo/slice(int).$(O) tests/gotgo/slice(list.List).$(O) tests/test(int).$(O) tests/test(string).$(O)

tests/gotgo/slice(int).$(O): tests/gotgo/slice(int).go

tests/gotgo/slice(list.List).$(O): tests/gotgo/slice(list.List).go tests/demo/list(int).$(O)

tests/test(int).$(O): tests/test(int).go

tests/test(string).$(O): tests/test(string).go

# found file tests/test.got to build...
# error:  tests/test.got:1:13: expected ';', found '('
tests/test.gotgo: tests/test.got

# ignoring tests/test.got.go, since it's a generated file
# ignoring tests/test.gotgo.go, since it's a generated file
installbins:  $(bindir)/gotgo $(bindir)/gotmake
installpkgs:  $(pkgdir)/gotgo/finalizer.gotgo $(pkgdir)/gotgo/slice.gotgo
