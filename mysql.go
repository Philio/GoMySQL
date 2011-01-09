// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

// Imports
import (
	"os"
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
	Errno int
	Error string
	
	// Logging
	Logging bool
	LogType uint8
	
	// Connection
	conn net.Conn
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

// Connect to server via socket
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
func (client *Client) error(errno int, error string) {
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
func (client *Client) Connect(netwk, raddr, user, passwd string, dbname ...string) (err os.Error) {
	// Connect to server
	client.conn, err = net.Dial(netwk, "", raddr)
	if err != nil {
		return
	}
	return
}
