# Copyright 2009 Dimiter Stanev, malkia@gmail.com
# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

MYBINFILES=gotit gotimports

all: $(MYBINFILES)

include $(GOROOT)/src/Make.$(GOARCH)

install:
	cp -f $(MYBINFILES) $(QUOTED_GOBIN)

.PHONY: test
.SUFFIXES: .$(O) .go .got .gotit

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotit:
	./gotit "$<"

got/gotit.$(O): got/buildit.$(O)
gotit.$(O): gotit.go got/gotit.$(O)
gotimports.$(O): gotimports.go got/gotit.$(O)

gotit: gotit.$(O)
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

tests/test.gotit: gotit
tests/demo/slice.gotit: gotit
tests/demo/list.gotit: gotit

tests/test(string).go: tests/test.gotit
	./tests/test.gotit string > "$@"
tests/test(int).go: tests/test.gotit
	./tests/test.gotit int > "$@"
tests/demo/list(int).go: tests/demo/list.gotit
	./tests/demo/list.gotit int > "$@"
tests/demo/slice(int).go: tests/demo/slice.gotit
	./tests/demo/slice.gotit int > "$@"
tests/demo/slice(list.List).go: tests/demo/slice.gotit tests/demo/list(int).$(O)
	./tests/demo/slice.gotit --import 'import list "./list(int)"' list.List > "$@"
