# Copyright 2009 Dimiter Stanev, malkia@gmail.com
# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

MYBINFILES=gotgo gotimports scripts/mkmake

all: $(MYBINFILES)

include $(GOROOT)/src/Make.$(GOARCH)

install:
	cp -f $(MYBINFILES) $(QUOTED_GOBIN)

.PHONY: test
.SUFFIXES: .$(O) .go .got .gotgo

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotgo:
	./gotgo "$<"

got/gotgo.$(O): got/buildit.$(O)
gotgo.$(O): gotgo.go got/gotgo.$(O)
gotimports.$(O): gotimports.go got/gotgo.$(O)

gotgo: gotgo.$(O)
	$(LD) -o $@ $<
gotimports: gotimports.$(O)
	$(LD) -o $@ $<

test: all tests/example
	./tests/example

tests/example.$(O): tests/example.go \
	tests/test(string).$(O) tests/test(int).$(O) \
	tests/demo/slice(list.List).$(O) \
	tests/demo/list(int).$(O) tests/demo/slice(int).$(O)
tests/example: tests/example.$(O)
	$(LD) -o $@ $<

tests/test.gotgo: gotgo
tests/demo/slice.gotgo: gotgo
tests/demo/list.gotgo: gotgo

tests/test(string).go: tests/test.gotgo
	./tests/test.gotgo string > "$@"
tests/test(int).go: tests/test.gotgo
	./tests/test.gotgo int > "$@"
tests/demo/list(int).go: tests/demo/list.gotgo
	./tests/demo/list.gotgo int > "$@"
tests/demo/slice(int).go: tests/demo/slice.gotgo
	./tests/demo/slice.gotgo int > "$@"
tests/demo/slice(list.List).go: tests/demo/slice.gotgo tests/demo/list(int).$(O)
	./tests/demo/slice.gotgo --import 'import list "./list(int)"' list.List > "$@"

scripts/mkmake.$(O): scripts/mkmake.go got/buildit.$(O)
scripts/mkmake: scripts/mkmake.$(O)
	$(LD) -o $@ $<
