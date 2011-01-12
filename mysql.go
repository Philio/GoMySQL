/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

import (
	"fmt"
	"net"
	"bufio"
	"io"
	"os"
	"log"
	"sync"
	"bytes"
)

const (
	Version       = "0.2.9"
	DefaultPort   = 3306
	DefaultSock   = "/var/run/mysqld/mysqld.sock"
	MaxPacketSize = 1 << 24
)

/**
 * The main MySQL struct
 */
type MySQL struct {
	Logging     bool
	Errno       int
	Error       string
	auth        *MySQLAuth
	conn        net.Conn
	reader      *bufio.Reader
	writer      *bufio.Writer
	sequence    uint8
	connected   bool
	serverInfo  *MySQLServerInfo
	curRes      *MySQLResult
	result      []*MySQLResult
	resultSaved bool
	pointer     int
	mutex       *sync.Mutex
}

/**
 * Server infomation
 */
type MySQLServerInfo struct {
	serverVersion   string
	protocolVersion uint8
	scrambleBuff    []byte
	capabilities    uint16
	language        uint8
	status          uint16
}

/**
 * Authentication infomation
 */
type MySQLAuth struct {
	host     string
	username string
	password string
	dbname   string
	port     int
	socket   string
}

/**
 * Create a new instance of the package
 */
func New() (mysql *MySQL) {
	// Create and return a new instance of MySQL
	mysql = new(MySQL)
	// Setup mutex
	mysql.mutex = new(sync.Mutex)
	return
}

/**
 * Connect to a server
 */
func (mysql *MySQL) Connect(params ...interface{}) (err os.Error) {
	if mysql.Logging {
		log.Print("Connect called")
	}
	// If already connected return
	if mysql.connected {
		err = os.NewError("Already connected to server")
		return
	}
	// Reset error/sequence vars
	mysql.reset()
	// Check min number of params
	if len(params) < 2 {
		err = os.NewError("A hostname and username are required to connect")
		return
	}
	// Parse params
	mysql.parseParams(params)
	// Connect to server
	err = mysql.connect()
	return
}

/**
 * Reconnect (if connection droppped etc)
 */
func (mysql *MySQL) Reconnect() (err os.Error) {
	if mysql.Logging {
		log.Print("Reconnect called")
	}
	// Check auth is set
	if mysql.auth == nil {
		err = os.NewError("Reconnect can only be called to re-establish a connection originally established by connect")
		return
	}
	// Close connection (force down)
	if mysql.connected {
		mysql.conn.Close()
		mysql.connected = false
	}
	// Reset error/sequence vars
	mysql.reset()
	// Call connect
	err = mysql.connect()
	return
}

/**
 * Close the connection to the server
 */
func (mysql *MySQL) Close() (err os.Error) {
	if mysql.Logging {
		log.Print("Close called")
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	// Send quit command
	err = mysql.command(COM_QUIT, "")
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence-1) + "] " + "Sent quit command to server")
	}
	// Close connection
	mysql.conn.Close()
	mysql.connected = false
	if mysql.Logging {
		log.Print("Closed connection to server")
	}
	return
}

/**
 * Perform SQL query
 */
func (mysql *MySQL) Query(sql string) (res *MySQLResult, err os.Error) {
	if mysql.Logging {
		if len(sql) > 512 {
			trim := sql[0:512]
			log.Print("Query called with SQL: " + trim + "...")
		} else {
			log.Print("Query called with SQL: " + sql)
		}
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	// Send query command
	err = mysql.command(COM_QUERY, sql)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence-1) + "] " + "Sent query command to server")
	}
	// Get result packet(s)
	for {
		// Get result packet
		err = mysql.getResult()
		if err != nil {
			return
		}
		// If result saved and buffer is empty break loop
		if mysql.resultSaved {
			break
		}
	}
	// If server sent result return it
	if len(mysql.result) > 0 {
		res = mysql.result[0]
		return
	} else {
		err = os.NewError("No valid result packets were received from MySQL")
	}
	return
}

/**
 * Perform SQL query with multiple result sets
 */
func (mysql *MySQL) MultiQuery(sql string) (res []*MySQLResult, err os.Error) {
	if mysql.Logging {
		if len(sql) > 512 {
			trim := sql[0:512]
			log.Print("MultiQuery called with SQL: " + trim + "...")
		} else {
			log.Print("MultiQuery called with SQL: " + sql)
		}
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	// Send query command
	err = mysql.command(COM_QUERY, sql)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence-1) + "] " + "Sent query command to server")
	}
	// Get result packet(s)
	for {
		// Get result packet
		err = mysql.getResult()
		if err != nil {
			return
		}
		// If result saved and buffer is empty break loop
		if mysql.resultSaved && mysql.reader.Buffered() == 0 {
			break
		}
	}
	// If server sent any results return them
	if len(mysql.result) > 0 {
		res = mysql.result
		return
	} else {
		err = os.NewError("No valid result packets were received from MySQL")
	}
	return
}

/**
 * Change database
 */
func (mysql *MySQL) ChangeDb(dbname string) (err os.Error) {
	if mysql.Logging {
		log.Print("ChangeDb called")
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	// Send command
	err = mysql.command(COM_INIT_DB, dbname)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence-1) + "] " + "Sent change db command to server")
	}
	// Get result packet
	err = mysql.getResult()
	return
}

/**
 * Ping server
 */
func (mysql *MySQL) Ping() (err os.Error) {
	if mysql.Logging {
		log.Print("Ping called")
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Lock mutex and defer unlock
	mysql.mutex.Lock()
	defer mysql.mutex.Unlock()
	// Reset error/sequence vars
	mysql.reset()
	// Send command
	err = mysql.command(COM_PING)
	if err != nil {
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence-1) + "] " + "Sent ping command to server")
	}
	// Get result packet
	err = mysql.getResult()
	return
}

/**
 * Escape string
 */
func (mysql *MySQL) Escape(str string) (esc string) {
	if mysql.Logging {
		log.Print("Escape called")
	}

	var prev byte
	var b bytes.Buffer

	for i := 0; i < len(str); i++ {
		switch str[i] {
		case '\'', '"':
			if prev != '\\' {
				b.WriteString(str[:i])
				b.WriteByte('\\')
				str = str[i:]
				i = 0
			}
		}

		prev = str[i]
	}

	b.WriteString(str)
	return b.String()
}

/**
 * Initialise a new statement
 */
func (mysql *MySQL) InitStmt() (stmt *MySQLStatement, err os.Error) {
	if mysql.Logging {
		log.Print("Initialise statement called")
	}
	// If not connected return
	if !mysql.connected {
		err = os.NewError("A connection to a MySQL server is required to use this function")
		return
	}
	// Create new statement and prepare query
	stmt = new(MySQLStatement)
	stmt.mysql = mysql
	return
}

/**
 * Clear error status, sequence number and result from previous command
 */
func (mysql *MySQL) reset() {
	mysql.Errno = 0
	mysql.Error = ""
	mysql.sequence = 0
	mysql.curRes = nil
	mysql.result = nil
	mysql.pointer = 0
}

/**
 * Parse params given to Connect()
 */
func (mysql *MySQL) parseParams(p []interface{}) {
	mysql.auth = new(MySQLAuth)
	// Assign default values
	mysql.auth.port = DefaultPort
	mysql.auth.socket = DefaultSock
	// Host / username are required
	mysql.auth.host = p[0].(string)
	mysql.auth.username = p[1].(string)
	// 3rd param should be a password
	if len(p) > 2 {
		mysql.auth.password = p[2].(string)
	}
	// 4th param should be a database name
	if len(p) > 3 {
		mysql.auth.dbname = p[3].(string)
	}
	// Reflect 5th param to determine if it is port or socket
	if len(p) > 4 {
		switch v := p[4].(type) {
			case int: mysql.auth.port = v
			case string: mysql.auth.socket = v
		}
	}
	return
}

/**
 * Create connection to server using unix socket or tcp/ip then setup buffered reader/writer
 */
func (mysql *MySQL) connect() (err os.Error) {
	// Connect via unix socket
	if mysql.auth.host == "localhost" || mysql.auth.host == "127.0.0.1" {
		mysql.conn, err = net.Dial("unix", "", mysql.auth.socket)
		// On error set the connect error details
		if err != nil {
			mysql.error(CR_CONNECTION_ERROR, fmt.Sprintf(CR_CONNECTION_ERROR_STR, mysql.auth.socket))
			return
		}
		if mysql.Logging {
			log.Print("Connected using unix socket")
		}
		// Connect via TCP
	} else {
		mysql.conn, err = net.Dial("tcp", "", fmt.Sprintf("%s:%d", mysql.auth.host, mysql.auth.port))
		// On error set the connect error details
		if err != nil {
			mysql.error(CR_CONN_HOST_ERROR, fmt.Sprintf(CR_CONN_HOST_ERROR_STR, mysql.auth.host, mysql.auth.port))
			return
		}
		if mysql.Logging {
			log.Print("Connected using TCP/IP")
		}
	}
	// Setup buffered reader and writer
	mysql.reader = bufio.NewReader(mysql.conn)
	mysql.writer = bufio.NewWriter(mysql.conn)
	// Get init packet from server
	err = mysql.init()
	if err != nil {
		return
	}
	// Send authenticate packet
	err = mysql.authenticate()
	if err != nil {
		return
	}
	// Get result packet
	err = mysql.getResult()
	if err != nil {
		return
	}
	mysql.connected = true
	return
}

/**
 * Read initial packet from server and populate server information
 */
func (mysql *MySQL) init() (err os.Error) {
	// Get header
	hdr := new(packetHeader)
	err = hdr.read(mysql.reader)
	// Check for read errors or incorrect sequence
	if err != nil || hdr.sequence != mysql.sequence {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR)
		return
	}
	// Get packet
	pkt := new(packetInit)
	pkt.header = hdr
	err = pkt.read(mysql.reader)
	if err != nil {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR)
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received init packet from server")
	}
	// Populate server info
	mysql.serverInfo = new(MySQLServerInfo)
	mysql.serverInfo.serverVersion = pkt.serverVersion
	mysql.serverInfo.protocolVersion = pkt.protocolVersion
	mysql.serverInfo.scrambleBuff = pkt.scrambleBuff
	mysql.serverInfo.capabilities = pkt.serverCaps
	mysql.serverInfo.language = pkt.serverLanguage
	mysql.serverInfo.status = pkt.serverStatus
	// Increment sequence
	mysql.sequence++
	return nil
}

/**
 * Send authentication packet to the server
 */
func (mysql *MySQL) authenticate() (err os.Error) {
	pkt := new(packetAuth)
	// Set client flags
	pkt.clientFlags = CLIENT_LONG_PASSWORD
	if len(mysql.auth.dbname) > 0 {
		pkt.clientFlags += CLIENT_CONNECT_WITH_DB
	}
	pkt.clientFlags += CLIENT_PROTOCOL_41
	pkt.clientFlags += CLIENT_TRANSACTIONS
	pkt.clientFlags += CLIENT_SECURE_CONN
	pkt.clientFlags += CLIENT_MULTI_STATEMENTS
	pkt.clientFlags += CLIENT_MULTI_RESULTS
	// Set max packet size
	pkt.maxPacketSize = MaxPacketSize
	// Set charset
	pkt.charsetNumber = mysql.serverInfo.language
	// Set username
	pkt.user = mysql.auth.username
	// Set password
	if len(mysql.auth.password) > 0 {
		// Encrypt password
		pkt.encrypt(mysql.auth.password, mysql.serverInfo.scrambleBuff)
	}
	// Set database name
	pkt.database = mysql.auth.dbname
	// Write packet
	err = pkt.write(mysql.writer)
	if err != nil {
		mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR)
		return
	}
	if mysql.Logging {
		log.Print("[" + fmt.Sprint(mysql.sequence) + "] Sent auth packet to server")
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Generic function to determine type of result packet received and process it
 */
func (mysql *MySQL) getResult() (err os.Error) {
	// Get header and validate header info
	hdr := new(packetHeader)
	err = hdr.read(mysql.reader)
	// Read error
	if err != nil {
		if mysql.connected {
			// Assume lost connection to server
			mysql.error(CR_SERVER_LOST, CR_SERVER_LOST_STR)
		} else {
			mysql.error(CR_SERVER_HANDSHAKE_ERR, CR_SERVER_HANDSHAKE_ERR_STR)
		}
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Check sequence number
	if hdr.sequence != mysql.sequence {
		mysql.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		return os.NewError("An error occured receiving packet from MySQL")
	}
	// Read the next byte to identify the type of packet
	c, err := mysql.reader.ReadByte()
	mysql.reader.UnreadByte()
	switch {
	// Unknown packet, remove it from the buffer
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
	case c == ResultPacketOK && mysql.curRes == nil:
		pkt := new(packetOK)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received ok packet from server")
		}
		// Create result
		mysql.curRes = new(MySQLResult)
		mysql.curRes.AffectedRows = pkt.affectedRows
		mysql.curRes.InsertId = pkt.insertId
		mysql.curRes.WarningCount = pkt.warningCount
		mysql.curRes.Message = pkt.message
		mysql.addResult()
	// Error Packet ff
	case c == ResultPacketError && mysql.curRes == nil:
		pkt := new(packetError)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		} else {
			mysql.error(int(pkt.errno), pkt.error)
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received error packet from server")
		}
		// Return error response
		err = os.NewError("An error was received from MySQL")
	// Result Set Packet 1-250 (first byte of Length-Coded Binary)
	case c >= 0x01 && c <= 0xfa && mysql.curRes == nil:
		pkt := new(packetResultSet)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received result set packet from server")
		}
		// Create result
		mysql.curRes = new(MySQLResult)
		mysql.curRes.FieldCount = pkt.fieldCount
		mysql.curRes.Fields = make([]*MySQLField, pkt.fieldCount)
		mysql.resultSaved = false
	// Field Packet 1-250 ("")
	case c >= 0x01 && c <= 0xfa && !mysql.curRes.fieldsEOF:
		pkt := new(packetField)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
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
		mysql.curRes.Fields[mysql.curRes.fieldsRead] = field
		// Increment fields read count
		mysql.curRes.fieldsRead++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received field packet from server")
		}
	// Row Data Packet 1-250 ("")
	case c >= 0x00 && c <= 0xfb && !mysql.curRes.rowsEOF:
		pkt := new(packetRowData)
		pkt.header = hdr
		pkt.fields = mysql.curRes.Fields
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		// Create row
		row := new(MySQLRow)
		row.Data = pkt.values
		if mysql.curRes.RowCount == 0 {
			mysql.curRes.Rows = make([]*MySQLRow, 1)
			mysql.curRes.Rows[0] = row
		} else {
			curRows := mysql.curRes.Rows
			mysql.curRes.Rows = make([]*MySQLRow, mysql.curRes.RowCount+1)
			copy(mysql.curRes.Rows, curRows)
			mysql.curRes.Rows[mysql.curRes.RowCount] = row
		}
		// Increment row count
		mysql.curRes.RowCount++
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received row data packet from server")
		}
	// EOF Packet fe
	case c == ResultPacketEOF:
		pkt := new(packetEOF)
		pkt.header = hdr
		err = pkt.read(mysql.reader)
		if err != nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if mysql.Logging {
			log.Print("[" + fmt.Sprint(mysql.sequence) + "] Received eof packet from server")
		}
		// Change EOF flag in result
		if mysql.curRes == nil {
			mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
			return
		}
		if !mysql.curRes.fieldsEOF {
			mysql.curRes.fieldsEOF = true
			if mysql.Logging {
				log.Print("End of field packets")
			}
		} else if !mysql.curRes.rowsEOF {
			mysql.curRes.rowsEOF = true
			if mysql.Logging {
				log.Print("End of row data packets")
			}
			mysql.addResult()
		}
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Add a temp result to the result array
 */
func (mysql *MySQL) addResult() {
	if mysql.pointer == 0 {
		mysql.result = make([]*MySQLResult, 1)
		mysql.result[0] = mysql.curRes
	} else {
		curRes := mysql.result
		mysql.result = make([]*MySQLResult, mysql.pointer+1)
		copy(mysql.result, curRes)
		mysql.result[mysql.pointer] = mysql.curRes
	}
	if mysql.Logging {
		log.Print("Current result set saved")
	}
	// Reset temp result
	mysql.curRes = nil
	mysql.resultSaved = true
	// Increment pointer
	mysql.pointer++
}

/**
 * Send a command to the server
 */
func (mysql *MySQL) command(command byte, args ...interface{}) (err os.Error) {
	pkt := new(packetCommand)
	pkt.command = command
	pkt.args = args
	err = pkt.write(mysql.writer)
	if err != nil {
		mysql.error(CR_MALFORMED_PACKET, CR_MALFORMED_PACKET_STR)
		return
	}
	// Increment sequence
	mysql.sequence++
	return
}

/**
 * Populate error variables
 */
func (mysql *MySQL) error(errno int, error string) {
	// Set err number/string
	mysql.Errno = errno
	mysql.Error = error
}
