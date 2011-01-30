// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

// Imports
import (
	"fmt"
	"log"
	"os"
	"net"
	"strings"
	"sync"
)

// Constants
const (
	// General
	VERSION          = "0.3.0-dev"
	DEFAULT_PORT     = "3306"
	DEFAULT_SOCKET   = "/var/run/mysqld/mysqld.sock"
	MAX_PACKET_SIZE  = 1<<24 - 1
	PROTOCOL_41      = 41
	PROTOCOL_40      = 40
	DEFAULT_PROTOCOL = PROTOCOL_41

	// Connection types
	TCP  = "tcp"
	UNIX = "unix"

	// Log methods
	LOG_SCREEN = 0x00
	LOG_FILE   = 0x01
)

// Client struct
type Client struct {
	// Errors
	Errno Errno
	Error Error

	// Logging
	LogLevel uint8
	LogType  uint8
	LogFile  *os.File

	// Credentials
	network string
	raddr   string
	user    string
	passwd  string
	dbname  string

	// Connection
	conn      net.Conn
	r         *reader
	w         *writer
	connected bool

	// Sequence
	protocol uint8
	sequence uint8

	// Server settings
	serverVersion  string
	serverProtocol uint8
	serverFlags    ClientFlag
	serverCharset  uint8
	serverStatus   ServerStatus
	scrambleBuff   []byte

	// Mutex for thread safety
	mutex sync.Mutex
}

// Create new client
func NewClient(protocol ...uint8) (c *Client) {
	if len(protocol) == 0 {
		protocol = make([]uint8, 1)
		protocol[0] = DEFAULT_PROTOCOL
	}
	c = &Client{
		protocol: protocol[0],
	}
	return
}

// Connect to server via TCP
func DialTCP(raddr, user, passwd string, dbname ...string) (c *Client, err os.Error) {
	c = NewClient(DEFAULT_PROTOCOL)
	// Add port if not set
	if strings.Index(raddr, ":") == -1 {
		raddr += ":" + DEFAULT_PORT
	}
	// Connect to server
	err = c.Connect(TCP, raddr, user, passwd, dbname...)
	return
}

// Connect to server via unix socket
func DialUnix(raddr, user, passwd string, dbname ...string) (c *Client, err os.Error) {
	c = NewClient(DEFAULT_PROTOCOL)
	// Use default socket if socket is empty
	if raddr == "" {
		raddr = DEFAULT_SOCKET
	}
	// Connect to server
	err = c.Connect(UNIX, raddr, user, passwd, dbname...)
	return
}

// Connect to the server
func (c *Client) Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Log connect
	c.log(1, "=== Begin connect ===")
	// Lock mutex/defer unlock
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Reset client
	c.reset()
	// Store connection credentials
	c.network = network
	c.raddr = raddr
	c.user = user
	c.passwd = passwd
	if len(dbname) > 0 {
		c.dbname = dbname[0]
	}
	// Call connect
	err = c.connect()
	if err != nil {
		return
	}
	// Set connected
	c.connected = true
	return
}

// Close connection to server
func (c *Client) Close() (err os.Error) {
	// Log close
	c.log(1, "=== Begin close ===")
	// Check connection
	if !c.connected {
		err = os.NewError("Must be connected to do this")
		return
	}
	// Lock mutex/defer unlock
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// Reset client
	c.reset()
	// Send close command
	c.command(COM_QUIT)
	// Close connection
	c.conn.Close()
	// Log disconnect
	c.log(1, "Disconnected")
	// Set connected
	c.connected = false
	return
}

// Change the current database
func (c *Client) ChangeDb(dbname string) (err os.Error) {
	return
}

// Send a query to the server
func (c *Client) Query(sql string) (err os.Error) {
	return
}

// Send multiple queries to the server
func (c *Client) MultiQuery(sql string) (err os.Error) {
	return
}

// Fetch all rows for a result and store it, returning the result set
func (c *Client) StoreResult() (result *Result, err os.Error) {
	return
}

// Use a result set, does not store rows
func (c *Client) UseResult() (result *Result, err os.Error) {
	return
}

// Check if more results are available
func (c *Client) MoreResults() (ok bool, err os.Error) {
	return
}

// Move to the next available result
func (c *Client) NextResult() (ok bool, err os.Error) {
	return
}

// Enable or disable autocommit
func (c *Client) AutoCommit(state bool) (err os.Error) {
	return
}

// Commit a transaction
func (c *Client) Commit() (err os.Error) {
	return
}

// Rollback a transaction
func (c *Client) Rollback() (err os.Error) {
	return
}

// Escape a string
func (c *Client) Escape(str string) (esc string) {
	return
}

// Initialise and prepare a new statement
func (c *Client) Prepare(sql string) (stmt *Statement, err os.Error) {
	return
}

// Initialise a new statment
func (c *Client) StmtInit() (stmt *Statement, err os.Error) {
	return
}

// Error handling
func (c *Client) error(errno Errno, error Error, args ...interface{}) {
	c.Errno = errno
	if len(args) > 0 {
		c.Error = Error(fmt.Sprintf(string(error), args...))
	} else {
		c.Error = error
	}
}

// Logging
func (c *Client) log(level uint8, format string, args ...interface{}) {
	// If logging is disabled, ignore
	if level > c.LogLevel {
		return
	}
	// Log based on logging type
	switch c.LogType {
	// Log to screen
	case LOG_SCREEN:
		log.Printf(format, args...)
	// Log to file
	case LOG_FILE:
		// If file pointer is nil return
		if c.LogFile == nil {
			return
		}
		// This is the same as log package does internally for logging
		// to the screen (via stderr) just requires an io.Writer
		l := log.New(c.LogFile, "", log.Ldate|log.Ltime)
		l.Printf(format, args...)
	// Not set
	default:
		return
	}
}

// Provide detailed log output for server capabilities
func (c *Client) logCaps() {
	c.log(3, "=== Server Capabilities ===")
	c.log(3, "Long password support: %d", c.serverFlags&CLIENT_LONG_PASSWORD)
	c.log(3, "Found rows: %d", c.serverFlags&CLIENT_FOUND_ROWS>>1)
	c.log(3, "All column flags: %d", c.serverFlags&CLIENT_LONG_FLAG>>2)
	c.log(3, "Connect with database support: %d", c.serverFlags&CLIENT_CONNECT_WITH_DB>>3)
	c.log(3, "No schema support: %d", c.serverFlags&CLIENT_NO_SCHEMA>>4)
	c.log(3, "Compression support: %d", c.serverFlags&CLIENT_COMPRESS>>5)
	c.log(3, "ODBC support: %d", c.serverFlags&CLIENT_ODBC>>6)
	c.log(3, "Load data local support: %d", c.serverFlags&CLIENT_LOCAL_FILES>>7)
	c.log(3, "Ignore spaces: %d", c.serverFlags&CLIENT_IGNORE_SPACE>>8)
	c.log(3, "4.1 protocol support: %d", c.serverFlags&CLIENT_PROTOCOL_41>>9)
	c.log(3, "Interactive client: %d", c.serverFlags&CLIENT_INTERACTIVE>>10)
	c.log(3, "Switch to SSL: %d", c.serverFlags&CLIENT_SSL>>11)
	c.log(3, "Ignore sigpipes: %d", c.serverFlags&CLIENT_IGNORE_SIGPIPE>>12)
	c.log(3, "Transaction support: %d", c.serverFlags&CLIENT_TRANSACTIONS>>13)
	c.log(3, "4.1 protocol authentication: %d", c.serverFlags&CLIENT_SECURE_CONN>>15)
}

// Provide detailed log output for the server status flags
func (c *Client) logStatus() {
	c.log(3, "=== Server Status ===")
	c.log(3, "In transaction: %d", c.serverStatus&SERVER_STATUS_IN_TRANS)
	c.log(3, "Auto commit enabled: %d", c.serverStatus&SERVER_STATUS_AUTOCOMMIT>>1)
	c.log(3, "More results exist: %d", c.serverStatus&SERVER_MORE_RESULTS_EXISTS>>3)
	c.log(3, "No good indexes were used: %d", c.serverStatus&SERVER_QUERY_NO_GOOD_INDEX_USED>>4)
	c.log(3, "No indexes were used: %d", c.serverStatus&SERVER_QUERY_NO_INDEX_USED>>5)
	c.log(3, "Cursor exists: %d", c.serverStatus&SERVER_STATUS_CURSOR_EXISTS>>6)
	c.log(3, "Last row has been sent: %d", c.serverStatus&SERVER_STATUS_LAST_ROW_SENT>>7)
	c.log(3, "Database dropped: %d", c.serverStatus&SERVER_STATUS_DB_DROPPED>>8)
	c.log(3, "No backslash escapes: %d", c.serverStatus&SERVER_STATUS_NO_BACKSLASH_ESCAPES>>9)
	c.log(3, "Metadata has changed: %d", c.serverStatus&SERVER_STATUS_METADATA_CHANGED>>10)
}

// Reset the client
func (c *Client) reset() {
	c.Errno = 0
	c.Error = ""
	c.sequence = 0
}

// Performs the actual connect
func (c *Client) connect() (err os.Error) {
	// Connect to server
	err = c.dial()
	if err != nil {
		return
	}
	// Read initial packet from server
	err = c.init()
	if err != nil {
		return
	}
	// Send auth packet to server
	c.sequence++
	err = c.auth()
	if err != nil {
		return
	}
	// Read auth result from server
	c.sequence++
	err = c.authResult()
	return
}

// Connect to server
func (c *Client) dial() (err os.Error) {
	// Log connect
	c.log(1, "Connecting to server via %s to %s", c.network, c.raddr)
	// Connect to server
	c.conn, err = net.Dial(c.network, "", c.raddr)
	if err != nil {
		// Store error state
		if c.network == UNIX {
			c.error(CR_CONNECTION_ERROR, CR_CONNECTION_ERROR_STR, c.raddr)
		}
		if c.network == TCP {
			c.error(CR_CONN_HOST_ERROR, CR_CONN_HOST_ERROR_STR, c.network, c.raddr)
		}
		// Log error
		c.log(1, err.String())
		return
	}
	// Log connect success
	c.log(1, "Connected to server")
	// Create reader and writer
	c.r = newReader(c.conn)
	c.w = newWriter(c.conn)
	// Set the reader default protocol
	c.r.protocol = c.protocol
	return
}

// Read initial packet from server
func (c *Client) init() (err os.Error) {
	// Log read packet
	c.log(1, "Reading handshake initialization packet from server")
	// Read packet
	p, err := c.r.readPacket(PACKET_INIT)
	if err != nil {
		return
	}
	err = c.checkSequence(p.(*packetInit).sequence)
	if err != nil {
		return
	}
	// Log success
	c.log(1, "Received handshake initialization packet")
	// Assign values
	c.serverVersion = p.(*packetInit).serverVersion
	c.serverProtocol = p.(*packetInit).protocolVersion
	c.serverFlags = ClientFlag(p.(*packetInit).serverCaps)
	c.serverCharset = p.(*packetInit).serverLanguage
	c.serverStatus = ServerStatus(p.(*packetInit).serverStatus)
	c.scrambleBuff = p.(*packetInit).scrambleBuff
	// Extended logging [level 2+]
	if c.LogLevel > 1 {
		// Log server info
		c.log(2, "Server version: %s", c.serverVersion)
		c.log(2, "Protocol version: %d", c.serverProtocol)
	}
	// Full logging [level 3]
	if c.LogLevel > 2 {
		c.logCaps()
		c.logStatus()
	}
	// If we're using 4.1 protocol and server doesn't support, drop to 4.0
	if c.protocol == PROTOCOL_41 && c.serverFlags&CLIENT_PROTOCOL_41 == 0 {
		c.protocol = PROTOCOL_40
		c.r.protocol = PROTOCOL_40
	}
	return
}

// Send auth packet to the server
func (c *Client) auth() (err os.Error) {
	// Log write packet
	c.log(1, "Sending authentication packet to server")
	// Construct packet
	p := &packetAuth{
		clientFlags:   uint32(CLIENT_MULTI_STATEMENTS | CLIENT_MULTI_RESULTS),
		maxPacketSize: MAX_PACKET_SIZE,
		charsetNumber: c.serverCharset,
		user:          c.user,
	}
	// Add protocol and sequence	// Full logging [level 3]
	if c.LogLevel > 2 {
		c.logStatus()
	}

	p.protocol = c.protocol
	p.sequence = c.sequence
	// Adjust client flags based on server support
	if c.serverFlags&CLIENT_LONG_PASSWORD > 0 {
		p.clientFlags |= uint32(CLIENT_LONG_PASSWORD)
	}
	if c.serverFlags&CLIENT_LONG_FLAG > 0 {
		p.clientFlags |= uint32(CLIENT_LONG_FLAG)
	}
	if c.serverFlags&CLIENT_TRANSACTIONS > 0 {
		p.clientFlags |= uint32(CLIENT_TRANSACTIONS)
	}
	// Check protocol
	if c.protocol == PROTOCOL_41 {
		p.clientFlags |= uint32(CLIENT_PROTOCOL_41 | CLIENT_SECURE_CONN)
		p.scrambleBuff = scramble41(c.scrambleBuff, []byte(c.passwd))
		// To specify a db name
		if c.serverFlags&CLIENT_CONNECT_WITH_DB > 0 && len(c.dbname) > 0 {
			p.clientFlags |= uint32(CLIENT_CONNECT_WITH_DB)
			p.database = c.dbname
		}
	} else {
		p.scrambleBuff = scramble323(c.scrambleBuff, []byte(c.passwd))
	}
	// Write packet
	err = c.w.writePacket(p)
	if err != nil {
		return
	}
	// Log write success
	c.log(1, "Sent authentication packet")
	return
}

// Send a command to the server
func (c *Client) command(command command, args ...interface{}) (err os.Error) {
	// Log write packet
	c.log(1, "Sending command packet to server")
	// Simple validation, arg count
	switch command {
	// No args
	case COM_QUIT, COM_STATISTICS, COM_PROCESS_INFO, COM_DEBUG, COM_PING:
		if len(args) != 0 {
			err = os.NewError(fmt.Sprintf("Invalid arg count, expected 0 but found %d", len(args)))
		}
	// 1 arg
	case COM_INIT_DB, COM_QUERY, COM_CREATE_DB, COM_DROP_DB, COM_REFRESH, COM_SHUTDOWN, COM_PROCESS_KILL, COM_STMT_PREPARE, COM_STMT_CLOSE, COM_STMT_RESET:
		if len(args) != 1 {
			err = os.NewError(fmt.Sprintf("Invalid arg count, expected 1 but found %d", len(args)))
		}
	// 2 args
	case COM_FIELD_LIST, COM_STMT_FETCH:
		if len(args) != 2 {
			err = os.NewError(fmt.Sprintf("Invalid arg count, expected 2 but found %d", len(args)))
		}
	// 4 args
	case COM_CHANGE_USER:
		if len(args) != 4 {
			err = os.NewError(fmt.Sprintf("Invalid arg count, expected 4 but found %d", len(args)))
		}
	// Commands with custom functions
	case COM_STMT_EXECUTE, COM_STMT_SEND_LONG_DATA:
		err = os.NewError("This command should not be used here")
	// Everything else e.g. replication unsupported
	default:
		err = os.NewError("This command is unsupported")
	}
	// Construct packet
	p := &packetCommand {
		command: command,
		args: args,
	}
	// Write packet
	err = c.w.writePacket(p)
	if err != nil {
		return
	}
	// Log write success
	c.log(1, "Sent command packet")
	return
}

// Get auth response
func (c *Client) authResult() (err os.Error) {
	// Log read result
	c.log(1, "Reading auth result packet from server")
	// Get result packet
	p, err := c.r.readPacket(PACKET_OK | PACKET_ERROR)
	if err != nil {
		return
	}
	// Process result packet
	switch i := p.(type) {
	case *packetOK:
		err = c.processOKResult(p.(*packetOK))
	case *packetError:
		err = c.processErrorResult(p.(*packetError))
	}
	return
}

// Sequence check
func (c *Client) checkSequence(sequence uint8) (err os.Error) {
	if sequence != c.sequence {
		c.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		c.log(1, "Sequence doesn't match, commands out of sync")
		err = os.NewError("Bad sequence number")
	}
	return
}

// Process OK packet
func (c *Client) processOKResult(p *packetOK) (err os.Error) {
	// Log OK result
	c.log(1, "Received OK packet")
	// Check sequence
	err = c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	c.serverStatus = ServerStatus(p.serverStatus)
	// Full logging [level 3]
	if c.LogLevel > 2 {
		c.logStatus()
	}
	return
}

// Process error packet
func (c *Client) processErrorResult(p *packetError) (err os.Error) {
	// Log error result
	c.log(1, "Received error packet")
	// Check sequence
	err = c.checkSequence(p.sequence)
	if err != nil {
		return
	}
	c.error(Errno(p.errno), Error(p.error))
	// Return error string as error
	err = os.NewError(p.error)
	return
}
