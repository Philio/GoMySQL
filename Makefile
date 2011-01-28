# GoMySQL - A MySQL client library for Go.
#
# Copyright 2010-2011 Phil Bayfield. All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.
include $(GOROOT)/src/Make.inc

TARG=github.com/Philio/GoMySQL
GOFILES=\
		const.go\
		error.go\
		mysql.go\
		packet.go\
		password.go\
		reader.go\
		result.go\
		statement.go\
		writer.go\

include $(GOROOT)/src/Make.pkg
