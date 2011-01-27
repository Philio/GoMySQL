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
)

// Packet reader struct
type reader struct {
	conn     net.Conn
	protocol uint8
}

// Create a new reader
func newReader(conn net.Conn) *reader {
	return &reader{
		conn:     conn,
		protocol: DEFAULT_PROTOCOL,
	}
}

// Read the next packet
func (r *reader) readPacket(types packetType) (p packetReadable, err os.Error) {
	// Read packet length
	pktLen, err := r.readNumber(3)
	if err != nil {
		return
	}
	// Read sequence
	pktSeq, err := r.readNumber(1)
	if err != nil {
		return
	}
	// Read rest of packet
	pktData := make([]byte, pktLen)
	nr, err := io.ReadFull(r.conn, pktData)
	if err != nil {
		return
	}
	if nr != int(pktLen) {
		err = os.NewError("Number of bytes read does not match packet length")
	}
	// Work out packet type
	switch {
	// Unknown packet
	default:
		err = os.NewError("Unknown/unexpected packet or packet type")
	// Initialisation / handshake packet, server > client
	case types&PACKET_INIT != 0:
		pi := new(packetInit)
		pi.protocol = r.protocol
		pi.sequence = uint8(pktSeq)
		return pi, pi.read(pktData)
	// Ok packet
	case types&PACKET_OK != 0 && pktData[0] == 0x00:
		pok := new(packetOK)
		pok.protocol = r.protocol
		pok.sequence = uint8(pktSeq)
		return pok, pok.read(pktData)
	// Error packet
	case types&PACKET_ERROR != 0 && pktData[0] == 0xff:
		per := new(packetError)
		per.protocol = r.protocol
		per.sequence = uint8(pktSeq)
		return per, per.read(pktData)
	}
	return
}

// Read n bytes long number
func (r *reader) readNumber(n uint8) (num uint64, err os.Error) {
	// Check max length
	if n > 8 {
		err = os.NewError("Cannot read a number greater than 8 bytes long")
		return
	}
	// Read bytes into array
	buf := make([]byte, n)
	nr, err := io.ReadFull(r.conn, buf)
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
