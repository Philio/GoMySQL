// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
)

// Prepared statement struct
type Statement struct {
	// Client pointer
	c *Client

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
	// Log query
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
	return
}

// Bind params
func (s *Statement) BindParams(params ...interface{}) (err os.Error) {
	// Check number of params is correct
	if len(params) != s.paramCount {
		return &ClientError{CR_INVALID_PARAMETER_NO, CR_INVALID_PARAMETER_NO_STR}
	}
	// Convert params into bytes
	return
}

// Send long data
func (s *Statement) SendLongData(num uint16, data string) (err os.Error) {
	return
}

// Execute
func (s *Statement) Execute() (err os.Error) {
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
