// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"net"
	"os"
)

// Packet writer struct
type writer struct {
	conn net.Conn
}

// Create a new reader
func newWriter(conn net.Conn) *writer {
	return &writer{
		conn: conn,
	}
}

// Write packet to the server
func (w *writer) writePacket(p packetWritable) (err os.Error) {
	// Get data in binary format
	pktData, err := p.write()
	if err != nil {
		return
	}
	// Write packet
	nw, err := w.conn.Write(pktData)
	if err != nil {
		return
	}
	if nw != len(pktData) {
		err = os.NewError("Number of bytes written does not match packet length")
	}
	return
}
