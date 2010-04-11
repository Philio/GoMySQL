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
	ParamCount	uint16
	
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
	}
	// Increment sequence
	stmt.mysql.sequence ++
}
