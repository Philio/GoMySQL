// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
)

// Packet type identifier
type packetType uint16

// Packet types
const (
	PACKET_INIT        packetType = 1 << iota
	PACKET_AUTH        packetType = 1 << iota
	PACKET_OK          packetType = 1 << iota
	PACKET_ERROR       packetType = 1 << iota
	PACKET_CMD         packetType = 1 << iota
	PACKET_RESULT      packetType = 1 << iota
	PACKET_FIELD       packetType = 1 << iota
	PACKET_ROW         packetType = 1 << iota
	PACKET_EOF         packetType = 1 << iota
	PACKET_OK_PREPARED packetType = 1 << iota
	PACKET_PARAM       packetType = 1 << iota
	PACKET_LONG_DATA   packetType = 1 << iota
	PACKET_EXECUTE     packetType = 1 << iota
	PACKET_ROW_BINARY  packetType = 1 << iota
)

// Readable packet interface
type packetReadable interface {
	read(data []byte) (err os.Error)
}

// Writable packet interface
type packetWriteable interface {
	write() (err os.Error)
}

// Generic packet interface (read/writable)
type packet interface {
	packetReadable
	packetWriteable
}

// Packet header
type header struct {
	length   uint32
	sequence uint8
}

// Init packet
type packetInit struct {
	*header
	protocolVersion uint8
	serverVersion   string
	threadId        uint32
	scrambleBuff    []byte
	serverCaps      uint16
	serverLanguage  uint8
	serverStatus    uint16
}

func (p *packetInit) read(data []byte) (err os.Error) {
	p.protocolVersion = data[0]
	return
}
