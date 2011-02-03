// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"io"
	"os"
)

// Packet writer struct
type writer struct {
	conn io.ReadWriteCloser
}

// Create a new reader
func newWriter(conn io.ReadWriteCloser) *writer {
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
