package mysql

import (
	"fmt"
	"os"
	"log"
)

/**
 * Prepared statement struct
 */
type MySQLStatement struct {
	StatementId	uint32

	Params		[]*MySQLParam
	ParamCount	uint16
	paramsRead	uint64
	paramsEOF	bool
	
	mysql		*MySQL
	MySQLResult
}

/**
 * Prepare sql statement
 */
func (stmt *MySQLStatement) Prepare(sql string) bool {
	if stmt.mysql.Logging { log.Stdout("Prepare statement called") }
	// Reset error/sequence vars
	stmt.mysql.reset()
	// Send command
	stmt.mysql.command(COM_STMT_PREPARE, sql)
	if stmt.mysql.Errno != 0 {
		return false
	}
	if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence - 1) + "] Sent prepare command to server") }
	// Get result packet(s)
	for {
		// Get result packet
		stmt.getResult()
		if stmt.mysql.Errno != 0 {
			return false
		}
		// If buffer is empty break loop
		if stmt.mysql.reader.Buffered() == 0 {
			break
		}
	}
	return true
}

/**
 * Execute statement
 */
func (stmt *MySQLStatement) Execute() {
	if stmt.mysql.Logging { log.Stdout("Execute statement called") }
	// Reset error/sequence vars
	stmt.mysql.reset()	
}

/**
 * Send long data packet
 */
func (stmt *MySQLStatement) SendLongData() {
	
}

/**
 * Close statement
 */
func (stmt *MySQLStatement) Close() bool {
	if stmt.mysql.Logging { log.Stdout("Close statement called") }
	// Reset error/sequence vars
	stmt.mysql.reset()
	// Send command
	stmt.mysql.command(COM_STMT_CLOSE, stmt.StatementId)
	if stmt.mysql.Errno != 0 {
		return false
	}
	if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence - 1) + "] Sent close statement command to server") }
	return true
}

/**
 * Reset statement
 */
func (stmt *MySQLStatement) Reset() bool {
	if stmt.mysql.Logging { log.Stdout("Reset statement called") }
	// Reset error/sequence vars
	stmt.mysql.reset()
	// Send command
	stmt.mysql.command(COM_STMT_RESET, stmt.StatementId)
	if stmt.mysql.Errno != 0 {
		return false
	}
	if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence - 1) + "] Sent reset statement command to server") }
	return true
}

/**
 * Function to read statement result packets
 */
func (stmt *MySQLStatement) getResult() {
	var err os.Error
	// Get header and validate header info
	hdr := new(packetHeader)
	err = hdr.read(stmt.mysql.reader)
	// Read error
	if err != nil {
		// Assume lost connection to server
		stmt.mysql.error(CR_SERVER_LOST, CR_SERVER_LOST_STR, false)
		return
	}
	// Check data length
	if int(hdr.length) > stmt.mysql.reader.Buffered() {
		stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
		return
	}
	// Check sequence number
	if hdr.sequence != stmt.mysql.sequence {
		stmt.mysql.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR, false)
		return
	}
	// Read the next byte to identify the type of packet
	c, err := stmt.mysql.reader.ReadByte()
	stmt.mysql.reader.UnreadByte()
	switch {
		// Unknown packet, read it and leave it for now
		default:
			bytes := make([]byte, hdr.length)
			stmt.mysql.reader.Read(bytes)
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received unknown packet from server with first byte: " + fmt.Sprint(c)) }
		// OK Packet 00
		case c == ResultPacketOK:
			pkt := new(packetOKPrepared)
			pkt.header = hdr
			err = pkt.read(stmt.mysql.reader)
			if err != nil {
				stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
			}
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received ok for prepared statement packet from server") }
			// Save statement info
			stmt.StatementId  = pkt.statementId
			stmt.FieldCount   = uint64(pkt.columnCount)
			stmt.ParamCount   = pkt.paramCount
			stmt.WarningCount = pkt.warningCount
			// Initialise params/fields
			stmt.Params = make([]*MySQLParam, pkt.paramCount)
			stmt.Fields = make([]*MySQLField, pkt.columnCount)
		// Error Packet ff
		case c == ResultPacketError:
			pkt := new(packetError)
			pkt.header = hdr
			err = pkt.read(stmt.mysql.reader)
			if err != nil {
				stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
			} else {
				stmt.mysql.error(int(pkt.errno), pkt.error, false)
			}
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received error packet from server") }
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
			err = pkt.read(stmt.mysql.reader)
			if err != nil {
				stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
			}
			// Increment params read
			stmt.paramsRead ++
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received param packet from server (ignored)") }
		// Field packet
		case c >= 0x01 && c <= 0xfa && stmt.FieldCount > 0 && !stmt.fieldsEOF:
			pkt := new(packetField)
			pkt.header = hdr
			err = pkt.read(stmt.mysql.reader)
			if err != nil {
				stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
				return
			}
			// Populate field data (ommiting anything which doesnt seam useful at time of writing)
			field := new(MySQLField)
			field.Name	    = pkt.name
			field.Length	    = pkt.length
			field.Type	    = pkt.fieldType
			field.Decimals	    = pkt.decimals
			field.Flags 	    = new(MySQLFieldFlags)
			field.Flags.process(pkt.flags)
			stmt.Fields[stmt.fieldsRead] = field
			// Increment fields read count
			stmt.fieldsRead ++
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received field packet from server") }
		// EOF Packet fe
		case c == ResultPacketEOF:
			pkt := new(packetEOF)
			pkt.header = hdr
			err = pkt.read(stmt.mysql.reader)
			if err != nil {
				stmt.mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
				return
			}
			if stmt.mysql.Logging { log.Stdout("[" + fmt.Sprint(stmt.mysql.sequence) + "] Received eof packet from server") }
			// Change EOF flag
			if stmt.ParamCount > 0 && !stmt.paramsEOF {
				stmt.paramsEOF = true
				if stmt.mysql.Logging { log.Stdout("End of param packets") }
			} else if stmt.fieldsEOF != true {
				stmt.fieldsEOF = true
				if stmt.mysql.Logging { log.Stdout("End of field packets") }
			}
	}
	// Increment sequence
	stmt.mysql.sequence ++
}

/**
 * Param definition
 */
type MySQLParam struct {
	Type		[]byte
	Flags		uint16
	Decimals	uint8
	Length		uint32
}
