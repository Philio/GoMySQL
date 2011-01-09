// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

// Imports
import (
	"os"
	"strings"
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
)

// Client struct
type Client struct {
	Errno int
	Error string
	Debug bool
	mutex sync.Mutex
}

// Create new client
func NewClient() (client *Client) {
	client = &Client{}
	return
}

// Connect to server via TCP
func DialTCP(raddr, user, pass string, dbname ...string) (client *Client, err os.Error) {
	client = NewClient()
	// Add port if not set
	if strings.Index(raddr, ":") == -1 {
		raddr += ":" + DefaultSock
	}
	// Connect to server
	err = client.Connect(NetTCP, raddr, user, pass, dbname...)
	return
}

// Connect to server via socket
func DialSocket(raddr, user, pass string, dbname ...string) (client *Client, err os.Error) {
	client = NewClient()
	// Use default socket if socket is empty
	if raddr == "" {
		raddr = DefaultSock
	}
	// Connect to server
	err = client.Connect(NetUnix, raddr, user, pass, dbname...)
	return
}

// Connect to the server
func (client *Client) Connect(net, raddr, user, pass string, dbname ...string) (err os.Error) {
	return
}
