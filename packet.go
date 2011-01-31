// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"bytes"
	"fmt"
	"os"
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
	protocol uint8
	sequence uint8
}

// Read a slice from the data
func (p *packetBase) readSlice(data []byte, delim byte) (slice []byte, err os.Error) {
	pos := bytes.IndexByte(data, delim)
	if pos > -1 {
		slice = data[:pos]
	} else {
		slice = data
		err = os.EOF
	}
	return
}

// Read length coded binary
func (p *packetBase) readLengthCodedBinary(data []byte) (num uint64, n int, err os.Error) {
	switch {
	// 0-250 = value of first byte
	case data[0] <= 250:
		num = uint64(data[0])
		n = 1
		return
	// 251 column value = NULL
	case data[0] == 251:
		num = 0
		n = 1
		return
	// 252 following 2 = value of following 16-bit word
	case data[0] == 252:
		n = 3
	// 253 following 3 = value of following 24-bit word
	case data[0] == 253:
		n = 4
	// 254 following 8 = value of following 64-bit word
	case data[0] == 254:
		n = 9
	}
	// Check there are enough bytes
	if len(data) < n {
		err = os.EOF
		return
	}
	// Unpack number
	num = p.unpackNumber(data[1:n])
	return
}

// Convert byte array into a number
func (p *packetBase) unpackNumber(data []byte) (num uint64) {
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
	slice, err := p.readSlice(data[pos:], 0x00)
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
	if ClientFlag(p.serverCaps)&CLIENT_PROTOCOL_41 > 0 {
		p.scrambleBuff = append(p.scrambleBuff, data[pos:pos+12]...)
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
	// For MySQL 4.1+
	if p.protocol == PROTOCOL_41 {
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

// Ok packet struct
type packetOK struct {
	packetBase
	affectedRows uint64
	insertId     uint64
	serverStatus uint16
	warningCount uint16
	message      string
}

// OK packet reader
func (p *packetOK) read(data []byte) (err os.Error) {
	// Recover errors
	defer func() {
		if e := recover(); e != nil {
			err = os.NewError(fmt.Sprintf("%s", e))
		}
	}()
	// Position (skip first byte/field count)
	pos := 1
	// Affected rows [length coded binary]
	num, n, err := p.readLengthCodedBinary(data[pos:])
	if err != nil {
		return
	}
	p.affectedRows = num
	pos += n
	// Insert id [length coded binary]
	num, n, err = p.readLengthCodedBinary(data[pos:])
	if err != nil {
		return
	}
	p.insertId = num
	pos += n
	// Server status [16 bit uint]
	p.serverStatus = uint16(p.unpackNumber(data[pos : pos+2]))
	pos += 2
	// Warning (4.1 only) [16 bit uint]
	if p.protocol == PROTOCOL_41 {
		p.warningCount = uint16(p.unpackNumber(data[pos : pos+2]))
		pos += 2
	}
	// Message (optional) [string]
	if pos < len(data) {
		p.message = string(data[pos:])
	}
	return
}

// Error packet struct
type packetError struct {
	packetBase
	errno uint16
	state string
	error string
}

// Error packet reader
func (p *packetError) read(data []byte) (err os.Error) {
	// Recover errors
	defer func() {
		if e := recover(); e != nil {
			err = os.NewError(fmt.Sprintf("%s", e))
		}
	}()
	// Position
	pos := 1
	// Error number [16 bit uint]
	p.errno = uint16(p.unpackNumber(data[pos : pos+2]))
	pos += 2
	// State (4.1 only) [string]
	if p.protocol == PROTOCOL_41 {
		pos++
		p.state = string(data[pos : pos+5])
		pos += 5
	}
	// Message [string]
	p.error = string(data[pos:])
	return
}

// Command packet struct
type packetCommand struct {
	packetBase
	command command
	args    []interface{}
}

// Command packet writer
func (p *packetCommand) write() (data []byte, err os.Error) {
	// Recover errors (if wrong param type supplied)
	defer func() {
		if e := recover(); e != nil {
			err = os.NewError(fmt.Sprintf("%s", e))
		}
	}()
	// Make slice from command byte
	data = []byte{byte(p.command)}
	// Add args to requests
	switch p.command {
	// Commands with 1 arg unterminated string
	case COM_INIT_DB, COM_QUERY, COM_CREATE_DB, COM_DROP_DB, COM_STMT_PREPARE:
		data = append(data, []byte(p.args[0].(string))...)
	// Commands with 1 arg 32 bit uint
	case COM_PROCESS_KILL, COM_STMT_CLOSE, COM_STMT_RESET:
		data = append(data, p.packNumber(uint64(p.args[0].(uint32)), 4)...)
	// Field list command
	case COM_FIELD_LIST:
		// Table name
		data = append(data, []byte(p.args[0].(string))...)
		// Terminator
		data = append(data, 0x00)
		// Column name
		if len(p.args) > 1 {
			data = append(data, []byte(p.args[1].(string))...)
		}
	// Refresh command
	case COM_REFRESH:
		data = append(data, byte(p.args[0].(Refresh)))
	// Shutdown command
	case COM_SHUTDOWN:
		data = append(data, byte(p.args[0].(Shutdown)))
	// Change user command
	case COM_CHANGE_USER:
		// User
		data = append(data, []byte(p.args[0].(string))...)
		// Terminator
		data = append(data, 0x00)
		// Scramble length for 4.1
		if p.protocol == PROTOCOL_41 {
			data = append(data, byte(len(p.args[1].([]byte))))
		}
		// Scramble buffer
		if len(p.args[1].([]byte)) > 0 {
			data = append(data, p.args[1].([]byte)...)
		}
		// Temrminator for 3.23
		if p.protocol == PROTOCOL_40 {
			data = append(data, 0x00)
		}
		// Database name
		if len(p.args[2].(string)) > 0 {
			data = append(data, []byte(p.args[2].(string))...)
		}
		// Terminator
		data = append(data, 0x00)
		// Character set number (5.1.23+ needs testing with earlier versions)
		data = append(data, p.packNumber(uint64(p.args[3].(uint16)), 2)...)
	// Fetch statement command
	case COM_STMT_FETCH:
		// Statement id
		data = append(data, p.packNumber(uint64(p.args[0].(uint32)), 4)...)
		// Number of rows
		data = append(data, p.packNumber(uint64(p.args[1].(uint32)), 4)...)
	}
	return
}
