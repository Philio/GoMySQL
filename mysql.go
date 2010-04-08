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
	
	tmpres		*MySQLResult
	result		[]*MySQLResult
	pointer		int
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
func (mysql *MySQL) Connect(host, username, password, dbname string, port int, socket string) bool {
	if mysql.connected { return false }
	if mysql.Logging { log.Stdout("Connect called") }
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
func (mysql *MySQL) Close() bool {
	if !mysql.connected { return false }
	if mysql.Logging { log.Stdout("Close called") }
	// Reset error/sequence vars
	mysql.reset()
	// Send quit command
	mysql.command(COM_QUIT, "")
	if mysql.Errno != 0 {
		return false
	}
	if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence - 1) + "] " + "Sent quit command to server") }
	// Close connection
	mysql.conn.Close()
	mysql.connected = false
	if mysql.Logging { log.Stdout("Closed connection to server") }
	return true
}

/**
 * Perform SQL query
 * @todo multiple queries work, but resulting packets are not read correctly
 * and end up appended to the message field of the first OK packet
 */
func (mysql *MySQL) Query(sql string) *MySQLResult {
	if !mysql.connected { return nil }
	if mysql.Logging { log.Stdout("Query called") }
	// Reset error/sequence vars
	mysql.reset()
	// Send query command
	mysql.command(COM_QUERY, sql)
	if mysql.Errno != 0 {
		return nil
	}
	if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence - 1) + "] " + "Sent query command to server") }
	// Get result packet(s)
	for {
		// Get result packet
		mysql.getResult(false)
		if mysql.Errno != 0 {
			return nil
		}
		// If buffer is empty break loop
		if mysql.reader.Buffered() == 0 {
			break;
		}
	}
	return mysql.result[0]
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
	mysql.tmpres = nil
	mysql.result = nil
	mysql.pointer = 0
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
		if mysql.Logging { log.Stdout("Connected using unix socket") }
	// Connect via TCP
	} else {
		mysql.conn, err = net.Dial("tcp", "", fmt.Sprintf("%s:%d", host, port))
		// On error set the connect error details
		if err != nil {
			mysql.error(CR_CONN_HOST_ERROR, fmt.Sprintf(CR_CONN_HOST_ERROR_STR, host, port), true)
		}
		if mysql.Logging { log.Stdout("Connected using TCP/IP") }
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
	if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received init packet from server") }
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
	if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Sent auth packet to server") }
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
		default:
			if mysql.Logging { log.Stdout("Received unknown packet from server") }
		// OK Packet 00
		case c == ResultPacketOK:
			pkt := new(PacketOK)
			pkt.header = hdr
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received ok packet from server") }
			// Create result
			mysql.tmpres = new(MySQLResult)
			mysql.tmpres.AffectedRows = pkt.AffectedRows
			mysql.tmpres.InsertId 	  = pkt.InsertId
			mysql.tmpres.WarningCount = pkt.WarningCount
			mysql.tmpres.Message	  = pkt.Message
			mysql.addResult()
		// Error Packet ff
		case c == ResultPacketError:
			pkt := new(PacketError)
			pkt.header = hdr
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
			} else {
				mysql.error(int(pkt.Errno), pkt.Error, connect)
			}
			if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received error packet from server") }
		// Result Set Packet 1-250 (first byte of Length-Coded Binary)
		// Field Packet 1-250 ("")
		// Row Data Packet 1-250 ("")
		case c >= 0x01 && c <= 0xfa:
			switch {
				// If result = nil then this is result set packet
				case mysql.tmpres == nil:
					pkt := new(PacketResultSet)
					pkt.header = hdr
					err = pkt.read(mysql.reader)
					if err != nil {
						mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
						return
					}
					if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received result set packet from server") }
					// Create result
					mysql.tmpres = new(MySQLResult)
					mysql.tmpres.FieldCount = pkt.FieldCount
					mysql.tmpres.Fields     = make([]*MySQLField, pkt.FieldCount)
				// If fields EOF not reached then this is field packet
				case mysql.tmpres.fieldsEOF != true:
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
					mysql.tmpres.Fields[mysql.tmpres.fieldsRead] = field
					// Increment fields read count
					mysql.tmpres.fieldsRead ++
					if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received field packet from server") }
				// If rows EOF not reached then this is row packet
				case mysql.tmpres.rowsEOF != true:
					pkt := new(PacketRowData)
					pkt.header = hdr
					pkt.FieldCount = mysql.tmpres.FieldCount
					err = pkt.read(mysql.reader)
					if err != nil {
						mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
						return
					}
					// Create row
					row := new(MySQLRow)
					row.Data = pkt.Values
					if mysql.tmpres.RowCount == 0 {
						mysql.tmpres.Rows = make([]*MySQLRow, 1)
						mysql.tmpres.Rows[0] = row
					} else {
						curRows := mysql.tmpres.Rows
						mysql.tmpres.Rows = make([]*MySQLRow, mysql.tmpres.RowCount + 1)
						for key, val := range curRows {
							mysql.tmpres.Rows[key] = val
						}
						mysql.tmpres.Rows[mysql.tmpres.RowCount] = row
					}
					// Increment row count
					mysql.tmpres.RowCount ++
					if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received row data packet from server") }
			}
		// EOF Packet fe
		case c == ResultPacketEOF:
			pkt := new(PacketEOF)
			pkt.header = hdr
			err = pkt.read(mysql.reader)
			if err != nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if mysql.Logging { log.Stdout("[" + fmt.Sprint(mysql.sequence) + "] " + "Received eof packet from server") }
			// Change EOF flag in result
			if mysql.tmpres == nil {
				mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR, connect)
				return
			}
			if mysql.tmpres.fieldsEOF != true {
				mysql.tmpres.fieldsEOF = true
				if mysql.Logging { log.Stdout("End of field packets") }
			} else if mysql.tmpres.rowsEOF != true {
				mysql.tmpres.rowsEOF = true
				if mysql.Logging { log.Stdout("End of row data packets") }
				mysql.addResult()
			}
	}
	// Increment sequence
	mysql.sequence ++
}

/**
 * Add a temp result to the result array
 */
func (mysql *MySQL) addResult() {
	if mysql.pointer == 0 {
		mysql.result = make([]*MySQLResult, 1)
		mysql.result[0] = mysql.tmpres
	} else {
		curRes := mysql.result
		mysql.result = make([]*MySQLResult, mysql.pointer + 1)
		for key, val := range curRes {
			mysql.result[key] = val
		}
		mysql.result[mysql.pointer] = mysql.tmpres
	}
	if mysql.Logging { log.Stdout("Current result set saved") }
	// Reset temp result
	mysql.tmpres = nil
	// Increment pointer
	mysql.pointer ++
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
