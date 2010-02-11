# Copyright 2009 Dimiter Stanev, malkia@gmail.com
# Copyright 2010 David Roundy, roundyd@physics.oregonstate.edu.
# All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

all: install

include $(GOROOT)/src/Make.$(GOARCH)

TARG    = gotit
GOFILES = $(TARG).go

include $(GOROOT)/src/Make.cmd
