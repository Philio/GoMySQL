// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
	"reflect"
	"strconv"
)

// Prepared statement struct
type Statement struct {
	// Client pointer
	c *Client

	// Statement status flags
	prepared      bool
	paramsBound   bool
	paramsRebound bool

	// Statement id
	statementId uint32

	// Params
	paramCount uint16
	paramType  [][]byte
	paramData  [][]byte

	// Columns (fields)
	columnCount uint64

	// Result
	AffectedRows uint64
	LastInsertId uint64
	Warnings     uint16
	result       *Result
}

// Prepare new statement
func (s *Statement) Prepare(sql string) (err os.Error) {
	// Log prepare
	s.c.log(1, "=== Begin prepare '%s' ===", sql)
	// Pre-run checks
	if !s.c.checkConn() || s.c.checkResult() {
		return &ClientError{CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR}
	}
	// Reset client
	s.reset()
	// Send close command
	err = s.c.command(COM_STMT_PREPARE, sql)
	if err != nil {
		return
	}
	// Read result from server
	s.c.sequence++
	_, err = s.getResult(PACKET_PREPARE_OK | PACKET_ERROR)
	if err != nil {
		return
	}
	// Read param packets
	if s.paramCount > 0 {
		for {
			s.c.sequence++
			eof, err := s.getResult(PACKET_PARAM | PACKET_EOF)
			if err != nil {
				return
			}
			if eof {
				break
			}
		}
	}
	// Read field packets
	if s.columnCount > 0 {
		// Create a new result
		for {
			s.c.sequence++
			eof, err := s.getResult(PACKET_FIELD | PACKET_EOF)
			if err != nil {
				return
			}
			if eof {
				break
			}
		}
	}
	// Statement is preapred
	s.prepared = true
	return
}

// Bind params
func (s *Statement) BindParams(params ...interface{}) (err os.Error) {
	// Check prepared
	if !s.prepared {
		return &ClientError{CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR}
	}
	// Check number of params is correct
	if len(params) != int(s.paramCount) {
		return &ClientError{CR_INVALID_PARAMETER_NO, CR_INVALID_PARAMETER_NO_STR}
	}
	// Convert params into bytes
	for num, param := range params {
		// Temp vars
		var t FieldType
		var d []byte
		// Switch on type
		switch param.(type) {
		// Nil
		case nil:
			t = FIELD_TYPE_NULL
		// Int
		case int:
			if strconv.IntSize == 32 {
				t = FIELD_TYPE_LONG
			} else {
				t = FIELD_TYPE_LONGLONG
			}
			d = itob(param.(int))
		// Uint
		case uint:
			if strconv.IntSize == 32 {
				t = FIELD_TYPE_LONG
			} else {
				t = FIELD_TYPE_LONGLONG
			}
			d = uitob(param.(uint))
		// Int8
		case int8:
			t = FIELD_TYPE_TINY
			d = []byte{byte(param.(int8))}
		// Uint8
		case uint8:
			t = FIELD_TYPE_TINY
			d = []byte{param.(uint8)}
		// Int16
		case int16:
			t = FIELD_TYPE_SHORT
			d = i16tob(param.(int16))
		// Uint16
		case uint16:
			t = FIELD_TYPE_SHORT
			d = ui16tob(param.(uint16))
		// Int32
		case int32:
			t = FIELD_TYPE_LONG
			d = i32tob(param.(int32))
		// Uint32
		case uint32:
			t = FIELD_TYPE_LONG
			d = ui32tob(param.(uint32))
		// Int64
		case int64:
			t = FIELD_TYPE_LONGLONG
			d = i64tob(param.(int64))
		// Uint64
		case uint64:
			t = FIELD_TYPE_LONGLONG
			d = ui64tob(param.(uint64))
		// Float32
		case float32:
			t = FIELD_TYPE_FLOAT
			d = f32tob(param.(float32))
		// Float64
		case float64:
			t = FIELD_TYPE_DOUBLE
			d = f64tob(param.(float64))
		// String
		case string:
			t = FIELD_TYPE_STRING
			d = []byte(param.(string))
		// Byte array
		case []byte:
			t = FIELD_TYPE_BLOB
			d = param.([]byte)
		// Other types
		default:
			return &ClientError{CR_UNSUPPORTED_PARAM_TYPE, s.c.fmtError(CR_UNSUPPORTED_PARAM_TYPE_STR, reflect.NewValue(param).Type(), num)}
		}
		// Append values
		s.paramType = append(s.paramType, []byte{byte(t), 0x0})
		s.paramData = append(s.paramData, d)
	}
	// Flag params as bound
	s.paramsBound = true
	s.paramsRebound = true
	return
}

// Send long data
func (s *Statement) SendLongData(num int, data []byte) (err os.Error) {
	// Log send long data
	s.c.log(1, "=== Begin send long data ===")
	// Check prepared
	if !s.prepared {
		return &ClientError{CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR}
	}
	// Pre-run checks
	if !s.c.checkConn() || s.c.checkResult() {
		return &ClientError{CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR}
	}
	// Reset client
	s.reset()
	// Data position (if data is longer than max packet length
	pos := 0
	// Send data
	for {
		// Construct packet
		p := &packetLongData{
			command:     uint8(COM_STMT_SEND_LONG_DATA),
			statementId: s.statementId,
			paramNumber: uint16(num),
		}
		// Add protocol and sequence
		p.protocol = s.c.protocol
		p.sequence = s.c.sequence
		// Add data
		if len(data[pos:]) > MAX_PACKET_SIZE-12 {
			p.data = data[pos : MAX_PACKET_SIZE-12]
			pos += MAX_PACKET_SIZE - 12
		} else {
			p.data = data[pos:]
			pos += len(data[pos:])
		}
		// Write packet
		err = s.c.w.writePacket(p)
		if err != nil {
			return
		}
		// Log write success
		s.c.log(1, "[%d] Sent long data packet", p.sequence)
		// Check if all data sent
		if pos == len(data) {
			break
		}
		// Increment sequence
		s.c.sequence++
	}
	return
}

// Execute
func (s *Statement) Execute() (err os.Error) {
	// Log execute
	s.c.log(1, "=== Begin execute ===")
	// Check prepared
	if !s.prepared {
		return &ClientError{CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR}
	}
	// Check params bound
	if s.paramCount > 0 && !s.paramsBound {
		return &ClientError{CR_PARAMS_NOT_BOUND, CR_PARAMS_NOT_BOUND_STR}
	}
	// Pre-run checks
	if !s.c.checkConn() || s.c.checkResult() {
		return &ClientError{CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR}
	}
	// Reset client
	s.reset()
	// Construct packet
	p := &packetExecute{
		command:        byte(COM_STMT_EXECUTE),
		statementId:    s.statementId,
		flags:          byte(CURSOR_TYPE_NO_CURSOR),
		iterationCount: 1,
		nullBitMap:     s.getNullBitMap(),
		paramType:      s.paramType,
		paramData:      s.paramData,
	}
	// Add protocol and sequence
	p.protocol = s.c.protocol
	p.sequence = s.c.sequence
	// Add rebound flag
	if s.paramsRebound {
		p.newParamsBound = byte(1)
	}
	// Write packet
	err = s.c.w.writePacket(p)
	if err != nil {
		return
	}
	// Log write success
	s.c.log(1, "[%d] Sent execute packet", p.sequence)
	return
}

// Bind result
func (s *Statement) BindResult(params ...interface{}) (err os.Error) {
	return
}

// Fetch next row 
func (s *Statement) Fetch() (err os.Error) {
	return
}

// Store result
func (s *Statement) StoreResult() (err os.Error) {
	return
}

// Free result
func (s *Statement) FreeResult() (err os.Error) {
	return
}

// Reset
func (s *Statement) Reset() (err os.Error) {
	return
}

// Close
func (s *Statement) Close() (err os.Error) {
	return
}

// Reset the statement
func (s *Statement) reset() {
	s.AffectedRows = 0
	s.LastInsertId = 0
	s.Warnings = 0
	s.result = nil
	s.c.reset()
}

// Get null bit map
func (s *Statement) getNullBitMap() (nbm []byte) {
	nbm = make([]byte, (s.paramCount+7)/8)
	bm := uint64(0)
	// Check if params are null (nil)
	for i := uint16(0); i < s.paramCount; i++ {
		if s.paramType[i][0] == byte(FIELD_TYPE_NULL) {
			bm += 1 << uint(i)
		}
	}
	// Convert the uint64 value into bytes
	for i := 0; i < len(nbm); i++ {
		nbm[i] = byte(bm >> uint(i*8))
	}
	return
}

// Get result
func (s *Statement) getResult(types packetType) (eof bool, err os.Error) {
	// Log read result
	s.c.log(1, "Reading result packet from server")
	// Get result packet
	p, err := s.c.r.readPacket(types)
	if err != nil {
		return
	}
	// Process result packet
	switch p.(type) {
	default:
		err = &ClientError{CR_UNKNOWN_ERROR, CR_UNKNOWN_ERROR_STR}
	case *packetPrepareOK:
		err = s.processPrepareOKResult(p.(*packetPrepareOK))
	case *packetError:
		err = s.c.processErrorResult(p.(*packetError))
	case *packetParameter:
		err = s.processParamResult(p.(*packetParameter))
	case *packetField:
		err = s.processFieldResult(p.(*packetField))
	case *packetEOF:
		eof = true
		err = s.processEOF(p.(*packetEOF))
	}
	return
}

// Process prepare OK result
func (s *Statement) processPrepareOKResult(p *packetPrepareOK) (err os.Error) {
	// Log result
	s.c.log(1, "[%d] Received prepare OK packet", p.sequence)
	// Check sequence
	err = s.c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	// Store packet data
	s.statementId = p.statementId
	s.paramCount = p.paramCount
	s.columnCount = uint64(p.columnCount)
	s.Warnings = p.warningCount
	return
}

// Process (ignore) parameter packet result
func (s *Statement) processParamResult(p *packetParameter) (err os.Error) {
	// Log result
	s.c.log(1, "[%d] Received parameter packet [ignored]", p.sequence)
	// Check sequence
	err = s.c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	return
}

// Process field packet result
func (s *Statement) processFieldResult(p *packetField) (err os.Error) {
	// Log result
	s.c.log(1, "[%d] Received field packet", p.sequence)
	// Check sequence
	err = s.c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	// Check if there is a result set
	if s.result == nil {
		return
	}
	// Assign fields if needed
	if len(s.result.fields) == 0 {
		s.result.fields = make([]*Field, s.result.fieldCount)
	}
	// Create new field and add to result
	s.result.fields[s.result.fieldPos] = &Field{
		Database: p.database,
		Table:    p.table,
		Name:     p.name,
		Length:   p.length,
		Type:     p.fieldType,
		Flags:    p.flags,
		Decimals: p.decimals,
	}
	s.result.fieldPos++
	return
	return
}

// Process EOF packet
func (s *Statement) processEOF(p *packetEOF) (err os.Error) {
	// Log EOF result
	s.c.log(1, "[%d] Received EOF packet", p.sequence)
	// Check sequence
	err = s.c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	// Store packet data
	if p.useStatus {
		s.c.serverStatus = ServerStatus(p.serverStatus)
		// Full logging [level 3]
		if s.c.LogLevel > 2 {
			s.c.logStatus()
		}
	}
	return
}
