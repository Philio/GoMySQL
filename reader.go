// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
	"net"
	"io"
	"bufio"
)

// Packet reader struct
type reader struct {
	br *bufio.Reader
}

// Create a new reader
func newReader(conn net.Conn) *reader {
	return &reader {
		br: bufio.NewReader(conn),
	}
}

// Read the next packet
func (r *reader) readPacket() (p *packetBin, err os.Error) {
	// Read packet length
	pLen, err := r.readNumber(3)
	if err != nil {
		return
	}
	// Read sequence
	pSeq, err := r.readNumber(1)
	if err != nil {
		return
	}
	// Create new packet
	p = &packetBin {
		length:   uint32(pLen),
		sequence: uint8(pSeq),
	}
	// Read rest of packet
	buf := make([]byte, p.length)
	nr, err := io.ReadFull(r.br, buf)
	if err == nil && nr != int(p.length) {
		err = os.NewError("Number of bytes read does not match packet length")
	}	
	return
}

// Read n bytes long number
func (r *reader) readNumber(n uint8) (num uint64, err os.Error) {
	// Check max length
	if n > 8 {
		err = os.NewError("Cannot read a number greater than 64 bits/8 bytes long")
		return
	}
	// Read bytes into array
	buf := make([]byte, n)
	nr, err := io.ReadFull(r.br, buf)
	if err != nil {
		return
	}
	if nr != int(n) {
		err = os.NewError("Number of bytes read does not match number of bytes requested")
		return
	}
	// Convert to uint64
	num = 0
	for i := uint8(0); i < n; i++ {
		num |= uint64(buf[i]) << (i * 8)
	}
	return
}
