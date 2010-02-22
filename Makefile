# Copyright 2009 Dimiter Stanev, malkia@gmail.com
# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

all: go gotit

include $(GOROOT)/src/Make.$(GOARCH)

.PHONY: test
.SUFFIXES: .$(O) .go .got .gotit

.go.$(O):
	cd `dirname "$<"`; $(GC) `basename "$<"`
.got.gotit:
	./gotit "$<"

gotit.$(O): gotit.go got/buildit.$(O)
go.$(O): go.go got/buildit.$(O)

gotit: gotit.$(O)
	$(LD) -o $@ $<
go: go.$(O)
	$(LD) -o $@ $<

test: all example
	./example

example.$(O): example.go tests/test(string).$(O) tests/test(int).$(O) \
	demo/list(int).$(O) demo/slice(int).$(O)
	$(GC) $<
example: example.$(O)
	$(LD) -o $@ $<

tests/test.gotit: gotit
demo/slice.gotit: gotit
demo/list.gotit: gotit

tests/test(string).go: tests/test.gotit
	./tests/test.gotit string > "$@"
tests/test(int).go: tests/test.gotit
	./tests/test.gotit int > "$@"
demo/list(int).go: demo/list.gotit
	./demo/list.gotit int > "$@"
demo/slice(int).go: demo/slice.gotit
	./demo/slice.gotit int > "$@"
