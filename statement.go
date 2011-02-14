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
	
	// Fields
	fieldCount uint16
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
	s.c.reset()
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
	switch i := p.(type) {
	default:
		err = &ClientError{CR_UNKNOWN_ERROR, CR_UNKNOWN_ERROR_STR}
	case *packetPrepareOK:
		err = s.processPrepareOKResult(p.(*packetPrepareOK))
	case *packetError:
		err = s.c.processErrorResult(p.(*packetError))
	case *packetParameter:
		err = s.processParamResult(p.(*packetParameter))
	case *packetEOF:
		eof = true
		err = s.c.processEOF(p.(*packetEOF))
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
	s.fieldCount = p.columnCount
	s.c.Warnings = p.warningCount
	return
}

// Process (ignore) parameter packet result
func (s *Statement) processParamResult(p *packetParameter) (err os.Error) {
	// Log result
	s.c.log(1, "[%d] Received parameter packet [ignored]", p.sequence)
	return
}
