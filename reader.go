// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"net"
	"bufio"
)

// Packet reader struct
type reader struct {
	rd *bufio.Reader
}

// Create a new reader
func newReader(conn net.Conn) *reader {
	return &reader {
		rd: bufio.NewReader(conn),
	}
}


