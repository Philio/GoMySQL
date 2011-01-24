// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
	"bytes"
	"fmt"
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

// Packet base struct
type packetBase struct {
	sequence uint8
}

// Read a slice from the data
func (p *packetBase) readSlice(data []byte, offset int, delim byte) (slice []byte, err os.Error) {
	 pos := bytes.IndexByte(data[offset:], delim)
	 if pos > -1 {
	 	slice = data[offset:pos+offset]
	 } else {
	 	slice = data[offset:]
	 	err = os.EOF
	 }
	 return
}

// Convert byte array into a number
func (p *packetBase) packNumber(data []byte) uint64 {
	num := uint64(0)
	for i := uint8(0); i < uint8(len(data)); i++ {
		num |= uint64(data[i]) << (i * 8)
	}
	return num
}

// Init packet
type packetInit struct {
	packetBase
	protocolVersion uint8
	serverVersion   string
	threadId        uint32
	scrambleBuff    []byte
	serverCaps      uint16
	serverLanguage  uint8
	serverStatus    uint16
}

// Init packet reader
func (p *packetInit) read(data []byte) (err os.Error) {
	// Recover errors
	defer func() {
		if e := recover(); e != nil {
			err = os.NewError(fmt.Sprintf("%s", e))
		}
	}()
	// Position
	pos := 0
	// Protocol version [8 bit uint]
	p.protocolVersion = data[pos]
	pos ++
	// Server version [null terminated string]
	slice, err := p.readSlice(data, pos, 0x00)
	if err != nil {
		return
	}
	p.serverVersion = string(slice)
	pos += len(slice) + 1
	// Thread id [32 bit uint]
	p.threadId = uint32(p.packNumber(data[pos:pos+4]))
	pos += 4
	// First part of scramble buffer [8 bytes]
	p.scrambleBuff = make([]byte, 8)
	p.scrambleBuff = data[pos:pos+8]
	pos += 9
	// Server capabilities [16 bit uint]
	p.serverCaps = uint16(p.packNumber(data[pos:pos+2]))
	pos += 2
	// Server language [8 bit uint]
	p.serverLanguage = data[pos]
	pos ++
	// Server status [16 bit uint]
	p.serverStatus = uint16(p.packNumber(data[pos:pos+2]))
	pos += 15
	// Second part of scramble buffer, if exists (4.1+) [13 bytes]
	if pos < len(data) {
		sBuff := p.scrambleBuff
		p.scrambleBuff = make([]byte, 20)
		copy(p.scrambleBuff[0:8], sBuff)
		copy(p.scrambleBuff[8:20], data[pos:])
	}
	return
}
