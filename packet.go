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
type packetType uint32

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
	PACKET_EOF_40      packetType = 1 << iota
	PACKET_EOF_41      packetType = 1 << iota
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
type packetWritable interface {
	write() (data []byte, err os.Error)
}

// Generic packet interface (read/writable)
type packet interface {
	packetReadable
	packetWritable
}

// Packet base struct
type packetBase struct {
	sequence uint8
}

// Read a slice from the data
func (p *packetBase) readSlice(data []byte, offset int, delim byte) (slice []byte, err os.Error) {
	pos := bytes.IndexByte(data[offset:], delim)
	if pos > -1 {
		slice = data[offset : pos+offset]
	} else {
		slice = data[offset:]
		err = os.EOF
	}
	return
}

// Convert byte array into a number
func (p *packetBase) unpackNumber(data []byte) (num uint64) {
	num = 0
	for i := uint8(0); i < uint8(len(data)); i++ {
		num |= uint64(data[i]) << (i * 8)
	}
	return
}

// Convert number into a byte array
func (p *packetBase) packNumber(num uint64, n uint8) (bytes []byte) {
	bytes = make([]byte, n)
	for i := uint8(0); i < n; i++ {
		bytes[i] = byte(num >> (i * 8))
	}
	return
}

// Prepend packet data with header info
func (p *packetBase) addHeader(data []byte) (pkt []byte) {
	pkt = p.packNumber(uint64(len(data)), 3)
	pkt = append(pkt, p.sequence)
	pkt = append(pkt, data...)
	return
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
	pos++
	// Server version [null terminated string]
	slice, err := p.readSlice(data, pos, 0x00)
	if err != nil {
		return
	}
	p.serverVersion = string(slice)
	pos += len(slice) + 1
	// Thread id [32 bit uint]
	p.threadId = uint32(p.unpackNumber(data[pos : pos+4]))
	pos += 4
	// First part of scramble buffer [8 bytes]
	p.scrambleBuff = make([]byte, 8)
	p.scrambleBuff = data[pos : pos+8]
	pos += 9
	// Server capabilities [16 bit uint]
	p.serverCaps = uint16(p.unpackNumber(data[pos : pos+2]))
	pos += 2
	// Server language [8 bit uint]
	p.serverLanguage = data[pos]
	pos++
	// Server status [16 bit uint]
	p.serverStatus = uint16(p.unpackNumber(data[pos : pos+2]))
	pos += 15
	// Second part of scramble buffer, if exists (4.1+) [13 bytes]
	if ClientFlags(p.serverCaps)&CLIENT_PROTOCOL_41 > 0 && pos < len(data) {
		p.scrambleBuff = append(p.scrambleBuff, data[pos:]...)
	}
	return
}

// Auth packet
type packetAuth struct {
	packetBase
	clientFlags   uint32
	maxPacketSize uint32
	charsetNumber uint8
	user          string
	scrambleBuff  []byte
	database      string
}

// Auth packet writer
func (p *packetAuth) write() (data []byte, err os.Error) {
	// Recover errors
	defer func() {
		if e := recover(); e != nil {
			err = os.NewError(fmt.Sprintf("%s", e))
		}
	}()
	// For MySQL 4.1+
	if ClientFlags(p.clientFlags)&CLIENT_PROTOCOL_41 > 0 {
		// Client flags
		data = p.packNumber(uint64(p.clientFlags), 4)
		// Max packet size
		data = append(data, p.packNumber(uint64(p.maxPacketSize), 4)...)
		// Charset
		data = append(data, p.charsetNumber)
		// Filler
		data = append(data, make([]byte, 23)...)
		// User
		if len(p.user) > 0 {
			data = append(data, []byte(p.user)...)
		}
		// Terminator
		data = append(data, 0x00)
		// Scramble buffer
		data = append(data, byte(len(p.scrambleBuff)))
		if len(p.scrambleBuff) > 0 {
			data = append(data, p.scrambleBuff...)
		}
		// Database name
		if len(p.database) > 0 {
			data = append(data, []byte(p.database)...)
			// Terminator
			data = append(data, 0x00)
		}
	// For MySQL < 4.1
	} else {
		// Client flags
		data = p.packNumber(uint64(p.clientFlags), 2)
		// Max packet size
		data = append(data, p.packNumber(uint64(p.maxPacketSize), 3)...)
		// User
		if len(p.user) > 0 {
			data = append(data, []byte(p.user)...)
		}
		// Terminator
		data = append(data, 0x00)
		// Scramble buffer
		if len(p.scrambleBuff) > 0 {
			data = append(data, p.scrambleBuff...)
		}
		// Padding
		data = append(data, 0x00)
	}
	// Add the packet header
	data = p.addHeader(data)
	return
}
