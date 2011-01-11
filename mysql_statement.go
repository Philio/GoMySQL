/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

import (
	"fmt"
	"os"
	"log"
	"strconv"
	"io"
)

const (
	ParamLimit = 64
)

/**
 * Prepared statement struct
 */
type MySQLStatement struct {
	Errno int
	Error string

	mysql *MySQL

	prepared bool

	StatementId uint32

	Params        []*MySQLParam
	ParamCount    uint16
	paramsRead    uint64
	paramsEOF     bool
	paramData     []interface{}
	paramSentLong []bool
	paramsBound   bool
	paramsRebound bool

	result      *MySQLResult
	resExecuted bool
	resNorows   bool
}

/**
 * Param definition
 * Note - this is based on the information in the protocol description documnet
 * however as the packets are 'broken' these packets are actually being ignored
 * at some point this may be useful.
 */
type MySQLParam struct {
	Type     []byte
	Flags    uint16
	Decimals uint8
	Length   uint32
}

/**
 * Prepare sql statement
 */
func (stmt *MySQLStatement) Prepare(sql string) (err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Prepare statement called")
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	stmt.reset()
	// Send command
	err = stmt.command(COM_STMT_PREPARE, sql)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + strconv.Uitoa(uint(mysql.sequence-1)) + "] Sent prepare command to server")
	}
	// Get result packet(s)
	for {
		// Get result packet
		err = stmt.getPrepareResult()
		if err != nil {
			return
		}
		// Break when end of field packets reached
		if stmt.result.fieldsEOF {
			break
		}
	}
	stmt.prepared = true
	return
}

/**
 * Bind params
 */
func (stmt *MySQLStatement) BindParams(params ...interface{}) (err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Bind params called")
	}
	// Check statement has been prepared
	if !stmt.prepared {
		stmt.error(CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR)
		err = os.NewError("Statement must be prepared to use this function")
		return
	}
	// Check param count
	if uint16(len(params)) != stmt.ParamCount {
		err = os.NewError("Param count mismatch, expecting " + strconv.Uitoa(uint(stmt.ParamCount)) + ", got " + strconv.Uitoa(uint(len(params))))
		return
	}
	// Save params
	stmt.paramData = params
	stmt.paramsBound = true
	stmt.paramsRebound = true
	return
}

/**
 * Send long data packet
 */
func (stmt *MySQLStatement) SendLongData(num uint16, data string) (err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Send long data called")
	}
	// Check statement has been prepared
	if !stmt.prepared {
		stmt.error(CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR)
		err = os.NewError("Statement must be prepared to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	stmt.reset()
	// Construct packet
	pkt := new(packetLongData)
	pkt.sequence = mysql.sequence
	pkt.command = COM_STMT_SEND_LONG_DATA
	pkt.statementId = stmt.StatementId
	pkt.paramNumber = num
	pkt.data = data
	err = pkt.write(mysql.writer)
	if err != nil {
		stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		return
	}
	mysql.sequence++
	if mysql.Logging {
		log.Print("[" + strconv.Uitoa(uint(mysql.sequence-1)) + "] " + "Sent long data packet to server")
	}
	return
}

/**
 * Execute statement
 */
func (stmt *MySQLStatement) Execute() (res *MySQLResult, err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Execute statement called")
	}
	// Check statement has been prepared
	if !stmt.prepared {
		stmt.error(CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR)
		err = os.NewError("Statement must be prepared to use this function")
		return
	}
	// Check params are bound
	if stmt.ParamCount > 0 && !stmt.paramsBound {
		stmt.error(CR_PARAMS_NOT_BOUND, CR_PARAMS_NOT_BOUND_STR)
		err = os.NewError("Params must be bound to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	stmt.reset()
	stmt.resExecuted = false
	// Construct packet
	pkt := new(packetExecute)
	pkt.command = COM_STMT_EXECUTE
	pkt.statementId = stmt.StatementId
	pkt.flags = CURSOR_TYPE_NO_CURSOR
	pkt.iterationCount = 1
	pkt.encodeNullBits(stmt.paramData)
	pkt.encodeParams(stmt.paramData)
	if stmt.paramsRebound {
		pkt.newParamBound = 1
	} else {
		pkt.newParamBound = 0
	}
	err = pkt.write(mysql.writer)
	if err != nil {
		stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		return
	}
	mysql.sequence++
	if mysql.Logging {
		log.Print("[" + strconv.Uitoa(uint(mysql.sequence-1)) + "] " + "Sent execute statement to server")
	}
	// Get result packet(s)
	for {
		// Get result packet
		err = stmt.getExecuteResult()
		if err != nil {
			return
		}
		// If buffer is empty break loop
		if stmt.resExecuted && (stmt.resNorows || stmt.result.rowsEOF) {
			break
		}
	}
	stmt.paramsRebound = false
	res = stmt.result
	return
}

/**
 * Close statement
 */
func (stmt *MySQLStatement) Close() (err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Close statement called")
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Check statement has been prepared
	if !stmt.prepared {
		stmt.error(CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR)
		err = os.NewError("Statement must be prepared to use this function")
		return
	}
	// Reset error/sequence vars
	mysql.reset()
	stmt.reset()
	// Send command
	err = stmt.command(COM_STMT_CLOSE, stmt.StatementId)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + strconv.Uitoa(uint(mysql.sequence-1)) + "] Sent close statement command to server")
	}
	return
}

/**
 * Reset statement
 */
func (stmt *MySQLStatement) Reset() (err os.Error) {
	mysql := stmt.mysql
	if mysql.Logging {
		log.Print("Reset statement called")
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Check statement has been prepared
	if !stmt.prepared {
		stmt.error(CR_NO_PREPARE_STMT, CR_NO_PREPARE_STMT_STR)
		err = os.NewError("Statement must be prepared to use this function")
		return
	}
	// Reset error/sequence vars
	mysql.reset()
	stmt.reset()
	// Send command
	err = stmt.command(COM_STMT_RESET, stmt.StatementId)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + strconv.Uitoa(uint(mysql.sequence-1)) + "] Sent reset statement command to server")
	}
	err = stmt.getResetResult()
	if err != nil {
		return
	}
	stmt.paramsRebound = true
	return
}

/**
 * Clear error status
 */
func (stmt *MySQLStatement) reset() {
	stmt.Errno = 0
	stmt.Error = ""
}

/**
 * Function to read reset statment result packet
 */
func (stmt *MySQLStatement) getResetResult() (err os.Error) {
	mysql := stmt.mysql
	hdr := new(packetHeader)
	err = hdr.read(mysql.reader)
	if err != nil {
		stmt.error(CR_SERVER_LOST, CR_SERVER_LOST_STR)
		return os.NewError("An error occurred receiving packet from MySQL")
	}
	if hdr.sequence != mysql.sequence {
		stmt.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		return os.NewError("An error occurred receiving packet from MySQL")
	}
	c, err := mysql.reader.ReadByte()
	mysql.reader.UnreadByte()
	switch c {
	default:
		bytes := make([]byte, hdr.length)
		_, err = io.ReadFull(mysql.reader, bytes)
		// Set error response
		if err != nil {
			err = os.NewError("An unknown packet was received from MySQL, in addition an error occurred when attempting to read the packet from the buffer: " + err.String());
		} else {
			err = os.NewError("An unknown packet was received from MySQL")
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received unknown packet from server with first byte: " + fmt.Sprint(c))
		}
	// OK packet
	case ResultPacketOK:
		pkt := new(packetOK)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received ok for reset statement packet from server")
		}
	// Error packet
	case ResultPacketError:
		pkt := new(packetError)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received error packet from server")
		}
	}
	mysql.sequence++
	return
}

/**
 * Function to read prepare result packets
 */
func (stmt *MySQLStatement) getPrepareResult() (err os.Error) {
	mysql := stmt.mysql
	// Get header and validate header info
	hdr := new(packetHeader)
	err = hdr.read(mysql.reader)
	// Read error
	if err != nil {
		// Assume lost connection to server
		stmt.error(CR_SERVER_LOST, CR_SERVER_LOST_STR)
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Check sequence number
	if hdr.sequence != mysql.sequence {
		stmt.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Read the next byte to identify the type of packet
	c, err := mysql.reader.ReadByte()
	mysql.reader.UnreadByte()
	switch {
	// Unknown packet, read it and leave it for now
	default:
		bytes := make([]byte, hdr.length)
		_, err = io.ReadFull(mysql.reader, bytes)
		// Set error response
		if err != nil {
			err = os.NewError("An unknown packet was received from MySQL, in addition an error occurred when attempting to read the packet from the buffer: " + err.String());
		} else {
			err = os.NewError("An unknown packet was received from MySQL")
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received unknown packet from server with first byte: " + fmt.Sprint(c))
		}
	// OK Packet 00
	case c == ResultPacketOK:
		pkt := new(packetOKPrepared)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received ok for prepared statement packet from server")
		}
		// Save statement info
		stmt.result = new(MySQLResult)
		stmt.StatementId = pkt.statementId
		stmt.result.FieldCount = uint64(pkt.columnCount)
		stmt.ParamCount = pkt.paramCount
		stmt.result.WarningCount = pkt.warningCount
		// Initialise params/fields
		stmt.Params = make([]*MySQLParam, pkt.paramCount)
		stmt.paramData = make([]interface{}, pkt.paramCount)
		stmt.paramSentLong = make([]bool, pkt.paramCount)
		stmt.result.Fields = make([]*MySQLField, pkt.columnCount)
	// Error Packet ff
	case c == ResultPacketError:
		pkt := new(packetError)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		} else {
			stmt.error(int(pkt.errno), pkt.error)
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received error packet from server")
		}
		// Return error response
		err = os.NewError("An error was received from MySQL")
	// Making assumption that statement packets follow similar format to result packets
	// If param count > 0 then first will get parameter packets following EOF
	// After this should get standard field packets followed by EOF
	// Parameter packet
	case c >= 0x01 && c <= 0xfa && stmt.ParamCount > 0 && !stmt.paramsEOF:
		// This packet simply reads the number of bytes in the buffer per header length param
		// The packet specification for these packets is wrong also within MySQL code it states:
		// skip parameters data: we don't support it yet (in libmysql/libmysql.c)
		pkt := new(packetParameter)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		// Increment params read
		stmt.paramsRead++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received param packet from server (ignored)")
		}
	// Field packet
	case c >= 0x01 && c <= 0xfa && stmt.result.FieldCount > 0 && !stmt.result.fieldsEOF:
		pkt := new(packetField)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		// Populate field data (ommiting anything which doesnt seam useful at time of writing)
		field := new(MySQLField)
		field.Name = pkt.name
		field.Length = pkt.length
		field.Type = pkt.fieldType
		field.Decimals = pkt.decimals
		field.Flags = new(MySQLFieldFlags)
		field.Flags.process(pkt.flags)
		stmt.result.Fields[stmt.result.fieldsRead] = field
		// Increment fields read count
		stmt.result.fieldsRead++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received field packet from server")
		}
	// EOF Packet fe
	case c == ResultPacketEOF:
		pkt := new(packetEOF)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received eof packet from server")
		}
		// Change EOF flag
		if stmt.ParamCount > 0 && !stmt.paramsEOF {
			stmt.paramsEOF = true
			if mysql.Logging {
				log.Print("End of param packets")
			}
			if stmt.result.FieldCount == 0 {
				stmt.result.fieldsEOF = true
				if mysql.Logging {
					log.Print("End of field packets (automatic, no fields in result)")
				}
			}
		} else if !stmt.result.fieldsEOF {
			stmt.result.fieldsEOF = true
			if mysql.Logging {
				log.Print("End of field packets")
			}
		}
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Function to read execute result packets
 */
func (stmt *MySQLStatement) getExecuteResult() (err os.Error) {
	mysql := stmt.mysql
	// Get header and validate header info
	hdr := new(packetHeader)
	err = hdr.read(mysql.reader)
	// Read error
	if err != nil {
		// Assume lost connection to server
		stmt.error(CR_SERVER_LOST, CR_SERVER_LOST_STR)
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Check sequence number
	if hdr.sequence != mysql.sequence {
		stmt.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Read the next byte to identify the type of packet
	c, err := mysql.reader.ReadByte()
	mysql.reader.UnreadByte()
	switch {
	// Unknown packet, read it and leave it for now
	default:
		bytes := make([]byte, hdr.length)
		_, err = io.ReadFull(mysql.reader, bytes)
		// Set error response
		if err != nil {
			err = os.NewError("An unknown packet was received from MySQL, in addition an error occurred when attempting to read the packet from the buffer: " + err.String());
		} else {
			err = os.NewError("An unknown packet was received from MySQL")
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received unknown packet from server with first byte: " + fmt.Sprint(c))
		}
	// OK Packet 00
	case c == ResultPacketOK && !stmt.resExecuted:
		pkt := new(packetOK)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received ok packet from server")
		}
		// Create result
		stmt.result = new(MySQLResult)
		stmt.result.RowCount = 0
		stmt.result.AffectedRows = pkt.affectedRows
		stmt.result.InsertId = pkt.insertId
		stmt.result.WarningCount = pkt.warningCount
		stmt.result.Message = pkt.message
		stmt.resExecuted = true
		stmt.resNorows = true
	// Error Packet ff
	case c == ResultPacketError:
		pkt := new(packetError)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		} else {
			stmt.error(int(pkt.errno), pkt.error)
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received error packet from server")
		}
		// Return error response
		err = os.NewError("An error was received from MySQL")
	// Result Set Packet 1-250 (first byte of Length-Coded Binary)
	case c >= 0x01 && c <= 0xfa && !stmt.resExecuted:
		pkt := new(packetResultSet)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received result set packet from server")
		}
		stmt.result = new(MySQLResult)
		// If fields sent again re-read incase for some reason something changed
		if pkt.fieldCount > 0 {
			stmt.result.FieldCount = pkt.fieldCount
			stmt.result.Fields = make([]*MySQLField, pkt.fieldCount)
			stmt.result.fieldsRead = 0
			stmt.result.fieldsEOF = false
		}
		stmt.resExecuted = true
	// Field Packet 1-250 ("")
	case c >= 0x01 && c <= 0xfa && stmt.result.FieldCount > stmt.result.fieldsRead && !stmt.result.fieldsEOF:
		pkt := new(packetField)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		// Populate field data (ommiting anything which doesnt seam useful at time of writing)
		field := new(MySQLField)
		field.Name = pkt.name
		field.Length = pkt.length
		field.Type = pkt.fieldType
		field.Decimals = pkt.decimals
		field.Flags = new(MySQLFieldFlags)
		field.Flags.process(pkt.flags)
		stmt.result.Fields[stmt.result.fieldsRead] = field
		// Increment fields read count
		stmt.result.fieldsRead++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received field packet from server")
		}
	// Binary row packets appear to always start 00
	case c == ResultPacketOK && stmt.resExecuted:
		pkt := new(packetBinaryRowData)
		pkt.header = hdr
		pkt.fields = stmt.result.Fields
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		// Create row
		row := new(MySQLRow)
		row.Data = pkt.values
		if stmt.result.RowCount == 0 {
			stmt.result.Rows = make([]*MySQLRow, 1)
			stmt.result.Rows[0] = row
		} else {
			curRows := stmt.result.Rows
			stmt.result.Rows = make([]*MySQLRow, stmt.result.RowCount+1)
			copy(stmt.result.Rows, curRows)
			stmt.result.Rows[stmt.result.RowCount] = row
		}
		// Increment row count
		stmt.result.RowCount++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received binary row data packet from server")
		}
	// EOF Packet fe
	case c == ResultPacketEOF:
		pkt := new(packetEOF)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received eof packet from server")
		}
		// Change EOF flag
		if stmt.result.FieldCount > 0 && !stmt.result.fieldsEOF {
			stmt.result.fieldsEOF = true
			if mysql.Logging {
				log.Print("End of field packets")
			}
		} else if !stmt.result.rowsEOF {
			stmt.result.rowsEOF = true
			if mysql.Logging {
				log.Print("End of row packets")
			}
		}
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Send a command to the server
 */
func (stmt *MySQLStatement) command(command byte, args ...interface{}) (err os.Error) {
	mysql := stmt.mysql
	pkt := new(packetCommand)
	pkt.command = command
	pkt.args = args
	err = pkt.write(mysql.writer)
	if err != nil {
		stmt.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		return
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Populate error variables
 */
func (stmt *MySQLStatement) error(errno int, error string) {
	stmt.Errno = errno
	stmt.Error = error
}
