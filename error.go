// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

type Errno int

const (
	CR_UNKNOWN_ERROR        Errno = 2000
	CR_SOCKET_CREATE_ERROR  Errno = 2001
	CR_CONNECTION_ERROR     Errno = 2002
	CR_CONN_HOST_ERROR      Errno = 2003
	CR_IPSOCK_ERROR         Errno = 2004
	CR_UNKNOWN_HOST         Errno = 2005
	CR_SERVER_GONE_ERROR    Errno = 2006
	CR_SERVER_HANDSHAKE_ERR Errno = 2012
	CR_SERVER_LOST          Errno = 2013
	CR_COMMANDS_OUT_OF_SYNC Errno = 2014
	CR_MALFORMED_PACKET     Errno = 2027
	CR_NO_PREPARE_STMT      Errno = 2030
	CR_PARAMS_NOT_BOUND     Errno = 2031
)

type Error string

const (
	CR_UNKNOWN_ERROR_STR        Error = "Unknown MySQL error"
	CR_SOCKET_CREATE_ERROR_STR  Error = "Can't create UNIX socket (%d)"
	CR_CONNECTION_ERROR_STR     Error = "Can't connect to local MySQL server through socket '%s'"
	CR_CONN_HOST_ERROR_STR      Error = "Can't connect to MySQL server on '%s' (%d)"
	CR_IPSOCK_ERROR_STR         Error = "Can't create TCP/IP socket (%d)"
	CR_UNKNOWN_HOST_STR         Error = "Uknown MySQL server host '%s' (%d)"
	CR_SERVER_GONE_ERROR_STR    Error = "MySQL server has gone away"
	CR_SERVER_HANDSHAKE_ERR_STR Error = "Error in server handshake"
	CR_SERVER_LOST_STR          Error = "Lost connection to MySQL server during query"
	CR_COMMANDS_OUT_OF_SYNC_STR Error = "Commands out of sync; you can't run this command now"
	CR_MALFORMED_PACKET_STR     Error = "Malformed Packet"
	CR_NO_PREPARE_STMT_STR      Error = "Statement not prepared"
	CR_PARAMS_NOT_BOUND_STR     Error = "No data supplied for parameters in prepared statement"
)
