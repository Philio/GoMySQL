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
	VERSION         = "0.3.0-dev"
	DEFAULT_PORT    = 3306
	DEFAULT_SOCKET  = "/var/run/mysqld/mysqld.sock"
	MAX_PACKET_SIZE = 1 << 24 - 1

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
	rd   *reader
	wr   *writer

	// Sequence
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
func NewClient() (cl *Client) {
	cl = &Client{}
	return
}

// Connect to server via TCP
func DialTCP(raddr, user, passwd string, dbname ...string) (cl *Client, err os.Error) {
	cl = NewClient()
	// Add port if not set
	if strings.Index(raddr, ":") == -1 {
		raddr += ":" + fmt.Sprintf("%d", DEFAULT_PORT)
	}
	// Connect to server
	err = cl.Connect(TCP, raddr, user, passwd, dbname...)
	return
}

// Connect to server via unix socket
func DialUnix(raddr, user, passwd string, dbname ...string) (cl *Client, err os.Error) {
	cl = NewClient()
	// Use default socket if socket is empty
	if raddr == "" {
		raddr = DEFAULT_SOCKET
	}
	// Connect to server
	err = cl.Connect(UNIX, raddr, user, passwd, dbname...)
	return
}

// Error handling
func (cl *Client) error(errno Errno, error Error) {
	cl.Errno = errno
	cl.Error = error
}

// Logging
func (cl *Client) log(level uint8, msg string) {
	// If logging is disabled, ignore
	if level > cl.LogLevel {
		return
	}
	// Log based on logging type
	switch cl.LogType {
	// Log to screen
	case LOG_SCREEN:
		log.Print(msg)
	// Log to file
	case LOG_FILE:
		// If file pointer is nil return
		if cl.LogFile == nil {
			return
		}
		// This is the same as log package does internally for logging
		// to the screen (via stderr) just requires an io.Writer
		l := log.New(cl.LogFile, "", log.Ldate|log.Ltime)
		l.Print(msg)
	// Not set
	default:
		return
	}
}

// Provide detailed log output for server capabilities
func (cl *Client) logCaps() {
	cl.log(3, fmt.Sprintf("Long password support: %d", cl.serverFlags&CLIENT_LONG_PASSWORD))
	cl.log(3, fmt.Sprintf("Found rows: %d", cl.serverFlags&CLIENT_FOUND_ROWS>>1))
	cl.log(3, fmt.Sprintf("All column flags: %d", cl.serverFlags&CLIENT_LONG_FLAG>>2))
	cl.log(3, fmt.Sprintf("Connect with database support: %d", cl.serverFlags&CLIENT_CONNECT_WITH_DB>>3))
	cl.log(3, fmt.Sprintf("No schema support: %d", cl.serverFlags&CLIENT_NO_SCHEMA>>4))
	cl.log(3, fmt.Sprintf("Compression support: %d", cl.serverFlags&CLIENT_COMPRESS>>5))
	cl.log(3, fmt.Sprintf("ODBC support: %d", cl.serverFlags&CLIENT_ODBC>>6))
	cl.log(3, fmt.Sprintf("Load data local support: %d", cl.serverFlags&CLIENT_LOCAL_FILES>>7))
	cl.log(3, fmt.Sprintf("Ignore spaces: %d", cl.serverFlags&CLIENT_IGNORE_SPACE>>8))
	cl.log(3, fmt.Sprintf("4.1 protocol support: %d", cl.serverFlags&CLIENT_PROTOCOL_41>>9))
	cl.log(3, fmt.Sprintf("Interactive client: %d", cl.serverFlags&CLIENT_INTERACTIVE>>10))
	cl.log(3, fmt.Sprintf("Switch to SSL: %d", cl.serverFlags&CLIENT_SSL>>11))
	cl.log(3, fmt.Sprintf("Ignore sigpipes: %d", cl.serverFlags&CLIENT_IGNORE_SIGPIPE>>12))
	cl.log(3, fmt.Sprintf("Transaction support: %d", cl.serverFlags&CLIENT_TRANSACTIONS>>13))
	cl.log(3, fmt.Sprintf("4.1 protocol authentication: %d", cl.serverFlags&CLIENT_SECURE_CONN>>15))
}

// Provide detailed log output for the server status flags
func (cl *Client) logStatus() {
	cl.log(3, fmt.Sprintf("In transaction: %d", cl.serverStatus&SERVER_STATUS_IN_TRANS))
	cl.log(3, fmt.Sprintf("Auto commit enabled: %d", cl.serverStatus&SERVER_STATUS_AUTOCOMMIT>>1))
	cl.log(3, fmt.Sprintf("More results exist: %d", cl.serverStatus&SERVER_MORE_RESULTS_EXISTS>>3))
	cl.log(3, fmt.Sprintf("No good indexes were used: %d", cl.serverStatus&SERVER_QUERY_NO_GOOD_INDEX_USED>>4))
	cl.log(3, fmt.Sprintf("No indexes were used: %d", cl.serverStatus&SERVER_QUERY_NO_INDEX_USED>>5))
	cl.log(3, fmt.Sprintf("Cursor exists: %d", cl.serverStatus&SERVER_STATUS_CURSOR_EXISTS>>6))
	cl.log(3, fmt.Sprintf("Last row has been sent: %d", cl.serverStatus&SERVER_STATUS_LAST_ROW_SENT>>7))
	cl.log(3, fmt.Sprintf("Database dropped: %d", cl.serverStatus&SERVER_STATUS_DB_DROPPED>>8))
	cl.log(3, fmt.Sprintf("No backslash escapes: %d", cl.serverStatus&SERVER_STATUS_NO_BACKSLASH_ESCAPES>>9))
	cl.log(3, fmt.Sprintf("Metadata has changed: %d", cl.serverStatus&SERVER_STATUS_METADATA_CHANGED>>10))
}

// Reset the client
func (cl *Client) reset() {
	cl.Errno = 0
	cl.Error = ""
	cl.sequence = 0
}

// Sequence check
func (cl *Client) checkSequence(sequence uint8) (err os.Error) {
	if sequence != cl.sequence {
		cl.error(CR_COMMANDS_OUT_OF_SYNC, CR_COMMANDS_OUT_OF_SYNC_STR)
		cl.log(1, "Sequence doesn't match, commands out of sync")
		err = os.NewError("Bad sequence number")
	}
	return
}

// Connect to the server
func (cl *Client) Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Reset client
	cl.reset()
	// Lock mutex/defer unlock
	cl.mutex.Lock()
	defer cl.mutex.Unlock()
	// Store connection credentials
	cl.network = network
	cl.raddr = raddr
	cl.user = user
	cl.passwd = passwd
	if len(dbname) > 0 {
		cl.dbname = dbname[0]
	}
	// Connect to server
	err = cl.dial()
	if err != nil {
		return
	}
	// Read initial packet from server
	err = cl.init()
	if err != nil {
		return
	}
	// Send auth packet to server
	cl.sequence++
	err = cl.auth()
	if err != nil {
		return
	}
	// Read response packet
	cl.sequence++
	p, err := cl.rd.readPacket(PACKET_OK|PACKET_ERROR)
	if err != nil {
		return
	}
	switch i := p.(type) {
	}
	return
}

// Connect to server
func (cl *Client) dial() (err os.Error) {
	// Log connect info
	cl.log(1, fmt.Sprintf("Connecting to server via %s to %s", cl.network, cl.raddr))
	// Connect to server
	cl.conn, err = net.Dial(cl.network, "", cl.raddr)
	if err != nil {
		// Store error state
		if cl.network == UNIX {
			cl.error(CR_CONNECTION_ERROR, Error(fmt.Sprintf(string(CR_CONNECTION_ERROR_STR), cl.raddr)))
		}
		if cl.network == TCP {
			parts := strings.Split(cl.raddr, ":", -1)
			if len(parts) == 2 {
				cl.error(CR_CONN_HOST_ERROR, Error(fmt.Sprintf(string(CR_CONN_HOST_ERROR_STR), parts[0], parts[1])))
			} else {
				cl.error(CR_UNKNOWN_ERROR, CR_UNKNOWN_ERROR_STR)
			}
		}
		// Log error
		cl.log(1, err.String())
		return
	}
	// Log connect success
	cl.log(1, "Connected to server")
	// Create reader and writer
	cl.rd = newReader(cl.conn)
	cl.wr = newWriter(cl.conn)
	return
}

// Read initial packet from server
func (cl *Client) init() (err os.Error) {
	// Read packet
	init, err := cl.rd.readPacket(PACKET_INIT)
	if err != nil {
		return
	}
	err = cl.checkSequence(init.(*packetInit).sequence)
	if err != nil {
		return
	}
	// Assign values
	cl.serverVersion = init.(*packetInit).serverVersion
	cl.serverProtocol = init.(*packetInit).protocolVersion
	cl.serverFlags = ClientFlags(init.(*packetInit).serverCaps)
	cl.serverCharset = init.(*packetInit).serverLanguage
	cl.serverStatus = ServerStatus(init.(*packetInit).serverStatus)
	cl.scrambleBuff = init.(*packetInit).scrambleBuff
	// Extended logging [level 2+]
	if cl.LogLevel > 1 {
		// Log server info
		cl.log(2, fmt.Sprintf("Server version: %s", cl.serverVersion))
		cl.log(2, fmt.Sprintf("Protocol version: %d", cl.serverProtocol))
	}
	// Full logging [level 3]
	if cl.LogLevel > 2 {
		cl.logCaps()
		cl.logStatus()
	}
	return
}

// Send auth packet to the server
func (cl *Client) auth() (err os.Error) {
	// Construct packet
	auth := &packetAuth{
		clientFlags:   uint32(CLIENT_MULTI_STATEMENTS | CLIENT_MULTI_RESULTS),
		maxPacketSize: MAX_PACKET_SIZE,
		charsetNumber: cl.serverCharset,
		user:          cl.user,
	}
	// Add sequence
	auth.sequence = cl.sequence
	// Adjust client flags based on server support
	if cl.serverFlags&CLIENT_LONG_PASSWORD > 0 {
		auth.clientFlags += uint32(CLIENT_LONG_PASSWORD)
	}
	if cl.serverFlags&CLIENT_LONG_FLAG > 0 {
		auth.clientFlags += uint32(CLIENT_LONG_FLAG)
	}
	if cl.serverFlags&CLIENT_TRANSACTIONS > 0 {
		auth.clientFlags += uint32(CLIENT_TRANSACTIONS)
	}
	if cl.serverFlags&(CLIENT_PROTOCOL_41|CLIENT_SECURE_CONN) > 0 {
		auth.clientFlags += uint32(CLIENT_PROTOCOL_41 | CLIENT_SECURE_CONN)
		auth.scrambleBuff = scramble41(cl.scrambleBuff, []byte(cl.passwd))
	} else {
		auth.scrambleBuff = scramble323(cl.scrambleBuff, []byte(cl.passwd))
	}
	// To specify a db name
	if cl.serverFlags&CLIENT_CONNECT_WITH_DB > 0 && len(cl.dbname) > 0 {
		auth.clientFlags += uint32(CLIENT_CONNECT_WITH_DB)
		auth.database = cl.dbname
	}
	// Write packet
	err = cl.wr.writePacket(auth)
	return
}

// Close connection to server
func (cl *Client) Close() (err os.Error) {
	return
}

// Change the current database
func (cl *Client) ChangeDb(dbname string) (err os.Error) {
	return
}

// Send a query to the server
func (cl *Client) Query(sql string) (err os.Error) {
	return
}

// Send multiple queries to the server
func (cl *Client) MultiQuery(sql string) (err os.Error) {
	return
}

// Fetch all rows for a result and store it, returning the result set
func (cl *Client) StoreResult() (result *Result, err os.Error) {
	return
}

// Use a result set, does not store rows
func (cl *Client) UseResult() (result *Result, err os.Error) {
	return
}

// Check if more results are available
func (cl *Client) MoreResults() (ok bool, err os.Error) {
	return
}

// Move to the next available result
func (cl *Client) NextResult() (ok bool, err os.Error) {
	return
}

// Enable or disable autocommit
func (cl *Client) AutoCommit(state bool) (err os.Error) {
	return
}

// Commit a transaction
func (cl *Client) Commit() (err os.Error) {
	return
}

// Rollback a transaction
func (cl *Client) Rollback() (err os.Error) {
	return
}

// Escape a string
func (cl *Client) Escape(str string) (esc string) {
	return
}

// Initialise and prepare a new statement
func (cl *Client) Prepare(sql string) (stmt *Statement, err os.Error) {
	return
}

// Initialise a new statment
func (cl *Client) StmtInit() (stmt *Statement, err os.Error) {
	return
}
