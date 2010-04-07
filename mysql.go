package mysql

import (
	"fmt"
	"net"
	"bufio"
	"os"
	"log"
)

const (
	DefaultPort 	= 3306
	DefaultSock 	= "/var/run/mysqld/mysqld.sock"
	MaxPacketSize	= 1 << 24
)

/**
 * The main MySQL struct
 */
type MySQL struct {
	Logging		bool

	ConnectErrno	int
	ConnectError	string

	Errno		int
	Error		string

	conn		net.Conn
	reader		*bufio.Reader
	writer		*bufio.Writer
	sequence	uint8
	connected	bool	
	
	serverInfo	*MySQLServerInfo
	
	result		*MySQLResult
}

/**
 * Server infomation
 */
type MySQLServerInfo struct {
	serverVersion	string
	protocolVersion	uint8
	scrambleBuff	[]byte
	capabilities	uint16
	language	uint8
	status		uint16
}

/**
 * Create a new instance of the package
 */
func New(logging bool) (mysql *MySQL) {
	// Create and return a new instance of MySQL
	mysql = new(MySQL)
	if (logging) {
		mysql.Logging = true
	}
	return
}

/**
 * Connect to a server
 */
func (mysql *MySQL) Connect(host, username, password, dbname string, port int, socket string) (connected bool) {
	if mysql == nil { return }
	if (mysql.Logging) { log.Stdout("Connect called") }
	// Reset error/sequence vars
	mysql.reset()
	// Connect to server
	if port == 0 {
		port = DefaultPort
	}
	if socket == "" {
		socket = DefaultSock
	}
	mysql.connect(host, port, socket)
	if mysql.ConnectErrno != 0 {
		return false
	}
	// Get init packet from server
	mysql.init()
	if mysql.ConnectErrno != 0 {
		return false
	}
	// Send authenticate packet
	mysql.authenticate(username, password, dbname)
	if mysql.ConnectErrno != 0 {
		return false
	}
	// Get result packet
	mysql.getResult(true)
	if mysql.ConnectErrno != 0 {
		return false
	}
	mysql.connected = true
	return true
}

/**
 * Close the connection to the server
 */
func (mysql *MySQL) Close() (closed bool) {
	if mysql == nil { return }
	if (mysql.Logging) { log.Stdout("Close called") }
	// Reset error/sequence vars
	mysql.reset()
	// Send quit command
	mysql.command(COM_QUIT, "")
	if mysql.Errno != 0 {
		return false
	}
	if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence - 1) + "] " + "Sent quit command to server") }
	// Close connection
	mysql.conn.Close()
	mysql.connected = false
	if (mysql.Logging) { log.Stdout("Closed connection to server") }
	return true
}

/**
 * Perform SQL query
 * @todo multiple queries work, but resulting packets are not read correctly
 * and end up appended to the message field of the first OK packet
 */
func (mysql *MySQL) Query(sql string) *MySQLResult {
	if mysql == nil { return nil }
	if (mysql.Logging) { log.Stdout("Query called") }
	// Reset error/sequence vars
	mysql.reset()
	// Send query command
	mysql.command(COM_QUERY, sql)
	if mysql.Errno != 0 {
		return nil
	}
	if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence - 1) + "] " + "Sent query command to server") }
	// Get result packet(s)
	mysql.getResult(false)
	if mysql.Errno != 0 {
		return nil
	}
	// Check if result set returned
	if mysql.result != nil {
		if mysql.result.FieldCount > 0 {
			// Read data
			for ;; {
				if mysql.result.rowsEOF != true {
					mysql.getResult(false)
					if mysql.Errno != 0 {
						return nil
					}
				} else {
					break
				}
			}
		}
	}
	return mysql.result
}

/**
 * Clear error status, sequence number and result from previous command
 */
func (mysql *MySQL) reset() {
	mysql.ConnectErrno = 0
	mysql.ConnectError = ""
	mysql.Errno = 0
	mysql.Error = ""
	mysql.sequence = 0
	mysql.result = nil
}

/**
 * Create connection to server using unix socket or tcp/ip then setup buffered reader/writer
 */
func (mysql *MySQL) connect(host string, port int, socket string) {
	var err os.Error
	// Connect via unix socket
	if host == "localhost" || host == "127.0.0.1" {
		mysql.conn, err = net.Dial("unix", "", socket);
		// On error set the connect error details
		if err != nil {
			mysql.error(CR_CONNECTION_ERROR, fmt.Sprintf(CR_CONNECTION_ERROR_STR, socket), true)
		}
		if (mysql.Logging) { log.Stdout("Connected using unix socket") }
	// Connect via TCP
	} else {
		mysql.conn, err = net.Dial("tcp", "", fmt.Sprintf("%s:%d", host, port))
		// On error set the connect error details
		if err != nil {
			mysql.error(CR_CONN_HOST_ERROR, fmt.Sprintf(CR_CONN_HOST_ERROR_STR, host, port), true)
		}
		if (mysql.Logging) { log.Stdout("Connected using TCP/IP") }
	}
	// Setup reader and writer
	mysql.reader = bufio.NewReader(mysql.conn)
	mysql.writer = bufio.NewWriter(mysql.conn)
}

/**
 * Read initial packet from server and populate server information
 */
func (mysql *MySQL) init() {
	var err os.Error
	// Get header
	hdr := new(PacketHeader)
	err = hdr.read(mysql.reader)
	// Check for read errors or incorrect sequence
	if err != nil || hdr.Sequence != mysql.sequence {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR, true)
		return
	}
	// Check read buffer size matches header length
	if int(hdr.Length) != mysql.reader.Buffered() {
		mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, true)
		return
	}
	// Get packet
	pkt := new(PacketInit)
	err = pkt.read(mysql.reader)
	if err != nil {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR, true)
		return
	}
	if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received init packet from server") }
	// Populate server info
	mysql.serverInfo = new(MySQLServerInfo)
	mysql.serverInfo.serverVersion	 = pkt.ServerVersion
	mysql.serverInfo.protocolVersion = pkt.ProtocolVersion
	mysql.serverInfo.scrambleBuff	 = pkt.ScrambleBuff
	mysql.serverInfo.capabilities	 = pkt.ServerCapabilities
	mysql.serverInfo.language	 = pkt.ServerLanguage
	mysql.serverInfo.status		 = pkt.ServerStatus
	// Increment sequence
	mysql.sequence ++
}

/**
 * Send authentication packet to the server
 */
func (mysql *MySQL) authenticate(username, password, dbname string) {
	var err os.Error
	pkt := new(PacketAuth)
	// Set client flags
	pkt.ClientFlags = CLIENT_LONG_PASSWORD
	if len(dbname) > 0 {
		pkt.ClientFlags += CLIENT_CONNECT_WITH_DB
	}
	pkt.ClientFlags += CLIENT_PROTOCOL_41
	pkt.ClientFlags += CLIENT_TRANSACTIONS
	pkt.ClientFlags += CLIENT_SECURE_CONNECTION
	pkt.ClientFlags += CLIENT_MULTI_STATEMENTS
	pkt.ClientFlags += CLIENT_MULTI_RESULTS
	// Set max packet size
	pkt.MaxPacketSize = MaxPacketSize
	// Set charset
	pkt.CharsetNumber = mysql.serverInfo.language
	// Set username 
	pkt.User = username
	// Set password
	if len(password) > 0 {
		// Encrypt password
		pkt.encrypt(password, mysql.serverInfo.scrambleBuff)
	}
	// Set database name
	pkt.DatabaseName = dbname
	// Write packet
	err = pkt.write(mysql.writer)
	if err != nil {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR, true)
	}
	if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Sent auth packet to server") }
	// Increment sequence
	mysql.sequence ++
}

/**
 * Generic function to determine type of result packet received and process it
 */
func (mysql *MySQL) getResult(connect bool) {
	var err os.Error
	// Get header and validate header info
	hdr := new(PacketHeader)
	err = hdr.read(mysql.reader)
	// Read error
	if err != nil {
		if connect {
			// Assume something screwed up with handshake
			mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR, true)
		} else {
			// Assume lost connection to server
			mysql.error(CR_SERVER_LOST, CR_SERVER_LOST_STR, false)
		}
		return
	}
	// Check data length
	if int(hdr.Length) > mysql.reader.Buffered() {
		mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
		return
	}
	// Check sequence number
	if hdr.Sequence != mysql.sequence {
		mysql.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR, connect)
		return
	}
	// Read the next byte to identify the type of packet
	c, err := mysql.reader.ReadByte()
	mysql.reader.UnreadByte()
	switch {
		// OK Packet 00
		case c == ResultPacketOK:
			pkt := new(PacketOK)
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received ok packet from server") }
			// Create result
			mysql.result = new(MySQLResult)
			mysql.result.AffectedRows = pkt.AffectedRows
			mysql.result.InsertId 	  = pkt.InsertId
			mysql.result.WarningCount = pkt.WarningCount
			mysql.result.Message	  = pkt.Message
		// Error Packet ff
		case c == ResultPacketError:
			pkt := new(PacketError)
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
			} else {
				mysql.error(int(pkt.Errno), pkt.Error, connect)
			}
			if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received error packet from server") }
		// Result Set Packet 1-250 (first byte of Length-Coded Binary)
		// Field Packet 1-250 ("")
		// Row Data Packet 1-250 ("")
		case c >= 0x01 && c <= 0xfa:
			switch {
				// If result = nil then this is result set packet
				case mysql.result == nil:
					pkt := new(PacketResultSet)
					pkt.header = hdr
					err = pkt.read(mysql.reader)
					if err != nil {
						mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
						return
					}
					if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received result set packet from server") }
					// Create result
					mysql.result = new(MySQLResult)
					mysql.result.FieldCount = pkt.FieldCount
					mysql.result.Fields = make([]*MySQLField, pkt.FieldCount)
				// If fields EOF not reached then this is field packet
				case mysql.result.fieldsEOF != true:
					pkt := new(PacketField)
					pkt.header = hdr
					err = pkt.read(mysql.reader)
					if err != nil {
						mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
						return
					}
					// Populate field data (ommiting anything which doesnt seam useful at time of writing)
					field := new(MySQLField)
					field.Name	    = pkt.Name
					field.Length	    = pkt.Length
					field.Type	    = pkt.Type
					field.Decimals	    = pkt.Decimals
					field.Flags 	    = new(MySQLFieldFlags)
					field.Flags.process(pkt.Flags)
					mysql.result.Fields[mysql.result.fieldsRead] = field
					// Increment fields read count
					mysql.result.fieldsRead ++
					if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received field packet from server") }
				// If rows EOF not reached then this is row packet
				case mysql.result.rowsEOF != true:
					pkt := new(PacketRowData)
					pkt.header = hdr
					pkt.FieldCount = mysql.result.FieldCount
					err = pkt.read(mysql.reader)
					if err != nil {
						mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
						return
					}
					// Create row
					row := new(MySQLRow)
					row.Data = pkt.Values
					if mysql.result.RowCount == 0 {
						mysql.result.Rows = make([]*MySQLRow, 1)
						mysql.result.Rows[0] = row
					} else {
						curRows := mysql.result.Rows
						mysql.result.Rows = make([]*MySQLRow, mysql.result.RowCount + 1)
						for key, val := range curRows {
							mysql.result.Rows[key] = val
						}
						mysql.result.Rows[mysql.result.RowCount] = row
					}
					// Increment row count
					mysql.result.RowCount ++
					if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received row data packet from server") }
			}
		// EOF Packet fe
		case c == ResultPacketEOF:
			pkt := new(PacketEOF)
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if (mysql.Logging) { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received eof packet from server") }
			// Change EOF flag in result
			if mysql.result == nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if mysql.result.fieldsEOF != true {
				mysql.result.fieldsEOF = true
				if (mysql.Logging) { log.Stdout("End of field packets") }
			} else if mysql.result.rowsEOF != true {
				mysql.result.rowsEOF = true
				if (mysql.Logging) { log.Stdout("End of row data packets") }
			}
	}
	// Increment sequence
	mysql.sequence ++
}

/**
 * Send a command to the server
 */
func (mysql *MySQL) command(command byte, arg string) {
	var err os.Error
	// Send command
	switch command {
		case COM_QUIT, COM_QUERY:
			pkt := new(PacketCommand)
			pkt.Command = command
			pkt.Arg = arg
			err = pkt.write(mysql.writer)
	}
	if err != nil {
		mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, false)
		return
	}
	// Increment sequence
	mysql.sequence ++
}

/**
 * Populate error variables
 */
func (mysql *MySQL) error(errno int, error string, connect bool) {
	if connect {
		mysql.ConnectErrno = errno
		mysql.ConnectError = error
	} else {
		mysql.Errno = errno
		mysql.Error = error
	}
}
