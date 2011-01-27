// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

// Imports
import (
	"os"
	"fmt"
	"log"
	"strings"
	"net"
	"sync"
)

// Constants
const (
	// General
	VERSION          = "0.3.0-dev"
	DEFAULT_PORT     = 3306
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
	conn net.Conn
	r    *reader
	w    *writer

	// Sequence
	protocol uint8
	sequence uint8

	// Server settings
	serverVersion  string
	serverProtocol uint8
	serverFlags    ClientFlags
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
		raddr += ":" + fmt.Sprintf("%d", DEFAULT_PORT)
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

// Error handling
func (c *Client) error(errno Errno, error Error) {
	c.Errno = errno
	c.Error = error
}

// Logging
func (c *Client) log(level uint8, msg string) {
	// If logging is disabled, ignore
	if level > c.LogLevel {
		return
	}
	// Log based on logging type
	switch c.LogType {
	// Log to screen
	case LOG_SCREEN:
		log.Print(msg)
	// Log to file
	case LOG_FILE:
		// If file pointer is nil return
		if c.LogFile == nil {
			return
		}
		// This is the same as log package does internally for logging
		// to the screen (via stderr) just requires an io.Writer
		l := log.New(c.LogFile, "", log.Ldate|log.Ltime)
		l.Print(msg)
	// Not set
	default:
		return
	}
}

// Provide detailed log output for server capabilities
func (c *Client) logCaps() {
	c.log(3, "=== Server Capabilities ===")
	c.log(3, fmt.Sprintf("Long password support: %d", c.serverFlags&CLIENT_LONG_PASSWORD))
	c.log(3, fmt.Sprintf("Found rows: %d", c.serverFlags&CLIENT_FOUND_ROWS>>1))
	c.log(3, fmt.Sprintf("All column flags: %d", c.serverFlags&CLIENT_LONG_FLAG>>2))
	c.log(3, fmt.Sprintf("Connect with database support: %d", c.serverFlags&CLIENT_CONNECT_WITH_DB>>3))
	c.log(3, fmt.Sprintf("No schema support: %d", c.serverFlags&CLIENT_NO_SCHEMA>>4))
	c.log(3, fmt.Sprintf("Compression support: %d", c.serverFlags&CLIENT_COMPRESS>>5))
	c.log(3, fmt.Sprintf("ODBC support: %d", c.serverFlags&CLIENT_ODBC>>6))
	c.log(3, fmt.Sprintf("Load data local support: %d", c.serverFlags&CLIENT_LOCAL_FILES>>7))
	c.log(3, fmt.Sprintf("Ignore spaces: %d", c.serverFlags&CLIENT_IGNORE_SPACE>>8))
	c.log(3, fmt.Sprintf("4.1 protocol support: %d", c.serverFlags&CLIENT_PROTOCOL_41>>9))
	c.log(3, fmt.Sprintf("Interactive client: %d", c.serverFlags&CLIENT_INTERACTIVE>>10))
	c.log(3, fmt.Sprintf("Switch to SSL: %d", c.serverFlags&CLIENT_SSL>>11))
	c.log(3, fmt.Sprintf("Ignore sigpipes: %d", c.serverFlags&CLIENT_IGNORE_SIGPIPE>>12))
	c.log(3, fmt.Sprintf("Transaction support: %d", c.serverFlags&CLIENT_TRANSACTIONS>>13))
	c.log(3, fmt.Sprintf("4.1 protocol authentication: %d", c.serverFlags&CLIENT_SECURE_CONN>>15))
}

// Provide detailed log output for the server status flags
func (c *Client) logStatus() {
	c.log(3, "=== Server Status ===")
	c.log(3, fmt.Sprintf("In transaction: %d", c.serverStatus&SERVER_STATUS_IN_TRANS))
	c.log(3, fmt.Sprintf("Auto commit enabled: %d", c.serverStatus&SERVER_STATUS_AUTOCOMMIT>>1))
	c.log(3, fmt.Sprintf("More results exist: %d", c.serverStatus&SERVER_MORE_RESULTS_EXISTS>>3))
	c.log(3, fmt.Sprintf("No good indexes were used: %d", c.serverStatus&SERVER_QUERY_NO_GOOD_INDEX_USED>>4))
	c.log(3, fmt.Sprintf("No indexes were used: %d", c.serverStatus&SERVER_QUERY_NO_INDEX_USED>>5))
	c.log(3, fmt.Sprintf("Cursor exists: %d", c.serverStatus&SERVER_STATUS_CURSOR_EXISTS>>6))
	c.log(3, fmt.Sprintf("Last row has been sent: %d", c.serverStatus&SERVER_STATUS_LAST_ROW_SENT>>7))
	c.log(3, fmt.Sprintf("Database dropped: %d", c.serverStatus&SERVER_STATUS_DB_DROPPED>>8))
	c.log(3, fmt.Sprintf("No backslash escapes: %d", c.serverStatus&SERVER_STATUS_NO_BACKSLASH_ESCAPES>>9))
	c.log(3, fmt.Sprintf("Metadata has changed: %d", c.serverStatus&SERVER_STATUS_METADATA_CHANGED>>10))
}

// Reset the client
func (c *Client) reset() {
	c.Errno = 0
	c.Error = ""
	c.sequence = 0
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

// Connect to the server
func (c *Client) Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Reset client
	c.reset()
	// Lock mutex/defer unlock
	c.mutex.Lock()
	defer c.mutex.Unlock()
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
	return
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
	// Log connect info
	c.log(1, fmt.Sprintf("Connecting to server via %s to %s", c.network, c.raddr))
	// Connect to server
	c.conn, err = net.Dial(c.network, "", c.raddr)
	if err != nil {
		// Store error state
		if c.network == UNIX {
			c.error(CR_CONNECTION_ERROR, Error(fmt.Sprintf(string(CR_CONNECTION_ERROR_STR), c.raddr)))
		}
		if c.network == TCP {
			parts := strings.Split(c.raddr, ":", -1)
			if len(parts) == 2 {
				c.error(CR_CONN_HOST_ERROR, Error(fmt.Sprintf(string(CR_CONN_HOST_ERROR_STR), parts[0], parts[1])))
			} else {
				c.error(CR_UNKNOWN_ERROR, CR_UNKNOWN_ERROR_STR)
			}
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
	init, err := c.r.readPacket(PACKET_INIT)
	if err != nil {
		return
	}
	err = c.checkSequence(init.(*packetInit).sequence)
	if err != nil {
		return
	}
	// Log success
	c.log(1, "Received handshake initialization packet")
	// Assign values
	c.serverVersion = init.(*packetInit).serverVersion
	c.serverProtocol = init.(*packetInit).protocolVersion
	c.serverFlags = ClientFlags(init.(*packetInit).serverCaps)
	c.serverCharset = init.(*packetInit).serverLanguage
	c.serverStatus = ServerStatus(init.(*packetInit).serverStatus)
	c.scrambleBuff = init.(*packetInit).scrambleBuff
	// Extended logging [level 2+]
	if c.LogLevel > 1 {
		// Log server info
		c.log(2, fmt.Sprintf("Server version: %s", c.serverVersion))
		c.log(2, fmt.Sprintf("Protocol version: %d", c.serverProtocol))
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
	auth := &packetAuth{
		clientFlags:   uint32(CLIENT_MULTI_STATEMENTS | CLIENT_MULTI_RESULTS),
		maxPacketSize: MAX_PACKET_SIZE,
		charsetNumber: c.serverCharset,
		user:          c.user,
	}
	// Add protocol and sequence	// Full logging [level 3]
	if c.LogLevel > 2 {
		c.logStatus()
	}

	auth.protocol = c.protocol
	auth.sequence = c.sequence
	// Adjust client flags based on server support
	if c.serverFlags&CLIENT_LONG_PASSWORD > 0 {
		auth.clientFlags |= uint32(CLIENT_LONG_PASSWORD)
	}
	if c.serverFlags&CLIENT_LONG_FLAG > 0 {
		auth.clientFlags |= uint32(CLIENT_LONG_FLAG)
	}
	if c.serverFlags&CLIENT_TRANSACTIONS > 0 {
		auth.clientFlags |= uint32(CLIENT_TRANSACTIONS)
	}
	// Check protocol
	if c.protocol == PROTOCOL_41 {
		auth.clientFlags |= uint32(CLIENT_PROTOCOL_41 | CLIENT_SECURE_CONN)
		auth.scrambleBuff = scramble41(c.scrambleBuff, []byte(c.passwd))
		// To specify a db name
		if c.serverFlags&CLIENT_CONNECT_WITH_DB > 0 && len(c.dbname) > 0 {
			auth.clientFlags |= uint32(CLIENT_CONNECT_WITH_DB)
			auth.database = c.dbname
		}
	} else {
		auth.scrambleBuff = scramble323(c.scrambleBuff, []byte(c.passwd))
	}
	// Write packet
	err = c.w.writePacket(auth)
	if err != nil {
		return
	}
	// Log write success
	c.log(1, "Sent authentication packet")
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
		// Check sequence
		err = c.checkSequence(p.(*packetOK).sequence)
		if err != nil {
			return
		}
		// Log OK result
		c.log(1, "Received OK packet")
		c.serverStatus = ServerStatus(p.(*packetOK).serverStatus)
		// Full logging [level 3]
		if c.LogLevel > 2 {
			c.logStatus()
		}
	case *packetError:
		// Check sequence
		err = c.checkSequence(p.(*packetError).sequence)
		if err != nil {
			return
		}
		// Log error result
		c.log(1, "Received error packet")
		c.error(Errno(p.(*packetError).errno), Error(p.(*packetError).error))
		err = os.NewError(p.(*packetError).error)
	}
	return
}

// Close connection to server
func (c *Client) Close() (err os.Error) {
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
