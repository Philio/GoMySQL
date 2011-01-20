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
	Version       = "0.3.0-dev"
	DefaultPort   = 3306
	DefaultSock   = "/var/run/mysqld/mysqld.sock"
	MaxPacketSize = 1 << 24
	NetTCP        = "tcp"
	NetUnix       = "unix"
	LogScreen     = 0x00
	LogFile       = 0x01
)

// Client struct
type Client struct {
	// Errors
	Errno Errno
	Error Error

	// Logging
	Logging bool
	LogType uint8
	LogFile *os.File

	// Connection
	conn   net.Conn
	rd     *reader
	wr     *writer

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
		raddr += ":" + DefaultSock
	}
	// Connect to server
	err = cl.Connect(NetTCP, raddr, user, passwd, dbname...)
	return
}

// Connect to server via unix socket
func DialUnix(raddr, user, passwd string, dbname ...string) (cl *Client, err os.Error) {
	cl = NewClient()
	// Use default socket if socket is empty
	if raddr == "" {
		raddr = DefaultSock
	}
	// Connect to server
	err = cl.Connect(NetUnix, raddr, user, passwd, dbname...)
	return
}

// Error handling
func (cl *Client) error(errno Errno, error Error) {
	cl.Errno = errno
	cl.Error = error
}

// Logging
func (cl *Client) log(msg string) {
	// If logging is disabled, ignore
	if !cl.Logging {
		return
	}
	// Log based on logging type
	switch cl.LogType {
	// Log to screen
	case LogScreen:
		log.Print(msg)
	// Log to file
	case LogFile:
		// If file pointer is nil return
		if cl.LogFile == nil {
			return
		}
		// This is the same as log package does internally for logging
		// to the screen (via stderr) just requires an io.Writer
		l := log.New(cl.LogFile, "", log.Ldate | log.Ltime)
		l.Print(msg)
	}
}

// Connect to the server
func (cl *Client) Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Connect to server
	cl.conn, err = net.Dial(network, "", raddr)
	if err != nil {
		// Store error state
		if network == NetUnix {
			cl.error(CR_CONNECTION_ERROR, Error(fmt.Sprintf(CR_CONNECTION_ERROR_STR, raddr)))
		}
		if network == NetTCP {
			parts := strings.Split(raddr, ":", -1)
			if len(parts) == 2 {
				cl.error(CR_CONN_HOST_ERROR, Error(fmt.Sprintf(CR_CONN_HOST_ERROR_STR, parts[0], parts[1])))
			} else {
				cl.error(CR_UNKNOWN_ERROR, CR_UNKNOWN_ERROR_STR)
			}
		}
		// Log error
		cl.log(err.String())
		return
	}
	// Create reader and writer
	cl.rd = newReader(cl.conn)
	cl.wr = newWriter(cl.conn)
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
