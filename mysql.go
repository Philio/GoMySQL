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
	"bufio"
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

	// Connection
	conn   net.Conn
	reader *bufio.Reader
	writer *bufio.Writer

	// Mutex for thread safety
	mutex sync.Mutex
}

// Create new client
func NewClient() (client *Client) {
	client = &Client{}
	return
}

// Connect to server via TCP
func DialTCP(raddr, user, passwd string, dbname ...string) (client *Client, err os.Error) {
	client = NewClient()
	// Add port if not set
	if strings.Index(raddr, ":") == -1 {
		raddr += ":" + DefaultSock
	}
	// Connect to server
	err = client.Connect(NetTCP, raddr, user, passwd, dbname...)
	return
}

// Connect to server via unix socket
func DialUnix(raddr, user, passwd string, dbname ...string) (client *Client, err os.Error) {
	client = NewClient()
	// Use default socket if socket is empty
	if raddr == "" {
		raddr = DefaultSock
	}
	// Connect to server
	err = client.Connect(NetUnix, raddr, user, passwd, dbname...)
	return
}

// Error handling
func (client *Client) error(errno Errno, error Error) {
	client.Errno = errno
	client.Error = error
}

// Logging
func (client *Client) log(msg string) {
	// If logging is disabled, ignore
	if !client.Logging {
		return
	}
	// Log based on logging type
	switch client.LogType {
	case LogScreen:
		log.Print(msg)
	}
}

// Connect to the server
func (client *Client) Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Connect to server
	client.conn, err = net.Dial(network, "", raddr)
	if err != nil {
		// Store error state
		if network == NetUnix {
			client.error(CR_CONNECTION_ERROR, Error(fmt.Sprintf(CR_CONNECTION_ERROR_STR, raddr)))
		}
		if network == NetTCP {
			parts := strings.Split(raddr, ":", -1)
			if len(parts) == 2 {
				client.error(CR_CONN_HOST_ERROR, Error(fmt.Sprintf(CR_CONN_HOST_ERROR_STR, parts[0], parts[1])))
			}
		}
		// Log error
		client.log(err.String())
		return
	}
	return
}

// Close connection to server
func (client *Client) Close() (err os.Error) {
	return
}

// Change the current database
func (client *Client) ChangeDb(dbname string) (err os.Error) {
	return
}

// Send a query to the server
func (client *Client) Query(sql string) (err os.Error) {
	return
}

// Send multiple queries to the server
func (client *Client) MultiQuery(sql string) (err os.Error) {
	return
}

// Fetch all rows for a result and store it, returning the result set
func (client *Client) StoreResult() (result *Result, err os.Error) {
	return
}

// Use a result set, does not store rows
func (client *Client) UseResult() (result *Result, err os.Error) {
	return
}

// Check if more results are available
func (client *Client) MoreResults() (ok bool, err os.Error) {
	return
}

// Move to the next available result
func (client *Client) NextResult() (ok bool, err os.Error) {
	return
}

// Enable or disable autocommit
func (client *Client) AutoCommit(state bool) (err os.Error) {
	return
}

// Commit a transaction
func (client *Client) Commit() (err os.Error) {
	return
}

// Rollback a transaction
func (client *Client) Rollback() (err os.Error) {
	return
}

// Escape a string
func (client *Client) Escape(str string) (esc string) {
	return
}

// Initialise and prepare a new statement
func (client *Client) Prepare(sql string) (stmt *Statement, err os.Error) {
	return
}

// Initialise a new statment
func (client *Client) StmtInit() (stmt *Statement, err os.Error) {
	return
}
