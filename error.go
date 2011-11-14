// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import "fmt"

// Client error types
type Errno int
type Errstr string

const (
	CR_UNKNOWN_ERROR               Errno = 2000
	CR_UNKNOWN_ERROR_STR           Errstr ="Unknown MySQL error"
	CR_SOCKET_CREATE_ERROR         Errno = 2001
	CR_SOCKET_CREATE_ERROR_STR     Errstr ="Can't create UNIX socket (%d)"
	CR_CONNECTION_ERROR            Errno = 2002
	CR_CONNECTION_ERROR_STR        Errstr ="Can't connect to local MySQL server through socket '%s'"
	CR_CONN_HOST_ERROR             Errno = 2003
	CR_CONN_HOST_ERROR_STR         Errstr ="Can't connect to MySQL server on '%s'"
	CR_IPSOCK_ERROR                Errno = 2004
	CR_IPSOCK_ERROR_STR            Errstr ="Can't create TCP/IP socket (%d)"
	CR_UNKNOWN_HOST                Errno = 2005
	CR_UNKNOWN_HOST_STR            Errstr ="Uknown MySQL server host '%s' (%d)"
	CR_SERVER_GONE_ERROR           Errno = 2006
	CR_SERVER_GONE_ERROR_STR       Errstr ="MySQL server has gone away"
	CR_VERSION_ERROR               Errno = 2007
	CR_VERSION_ERROR_STR           Errstr ="Protocol mismatch; server version = %d, client version = %d"
	CR_OUT_OF_MEMORY               Errno = 2008
	CR_OUT_OF_MEMORY_STR           Errstr ="MySQL client ran out of memory"
	CR_WRONG_HOST_INFO             Errno = 2009
	CR_WRONG_HOST_INFO_STR         Errstr ="Wrong host info"
	CR_LOCALHOST_CONNECTION        Errno = 2010
	CR_LOCALHOST_CONNECTION_STR    Errstr ="Localhost via UNIX socket"
	CR_TCP_CONNECTION              Errno = 2011
	CR_TCP_CONNECTION_STR          Errstr ="%s via TCP/IP"
	CR_SERVER_HANDSHAKE_ERR        Errno = 2012
	CR_SERVER_HANDSHAKE_ERR_STR    Errstr ="Error in server handshake"
	CR_SERVER_LOST                 Errno = 2013
	CR_SERVER_LOST_STR             Errstr ="Lost connection to MySQL server during query"
	CR_COMMANDS_OUT_OF_SYNC        Errno = 2014
	CR_COMMANDS_OUT_OF_SYNC_STR    Errstr ="Commands out of sync; you can't run this command now"
	CR_NAMEDPIPE_CONNECTION        Errno = 2015
	CR_NAMEDPIPE_CONNECTION_STR    Errstr ="Named pipe: %s"
	CR_NAMEDPIPEWAIT_ERROR         Errno = 2016
	CR_NAMEDPIPEWAIT_ERROR_STR     Errstr ="Can't wait for named pipe to host: %s pipe: %s (%lu)"
	CR_NAMEDPIPEOPEN_ERROR         Errno = 2017
	CR_NAMEDPIPEOPEN_ERROR_STR     Errstr ="Can't open named pipe to host: %s pipe: %s (%lu)"
	CR_NAMEDPIPESETSTATE_ERROR     Errno = 2018
	CR_NAMEDPIPESETSTATE_ERROR_STR Errstr ="Can't set state of named pipe to host: %s pipe: %s (%lu)"
	CR_CANT_READ_CHARSET           Errno = 2019
	CR_CANT_READ_CHARSET_STR       Errstr ="Can't initialize character set %s (path: %s)"
	CR_NET_PACKET_TOO_LARGE        Errno = 2020
	CR_NET_PACKET_TOO_LARGE_STR    Errstr ="Got packet bigger than 'max_allowed_packet' bytes"
	CR_EMBEDDED_CONNECTION         Errno = 2021
	CR_EMBEDDED_CONNECTION_STR     Errstr ="Embedded server"
	CR_PROBE_SLAVE_STATUS          Errno = 2022
	CR_PROBE_SLAVE_STATUS_STR      Errstr ="Error on SHOW SLAVE STATUS:"
	CR_PROBE_SLAVE_HOSTS           Errno = 2023
	CR_PROBE_SLAVE_HOSTS_STR       Errstr ="Error on SHOW SLAVE HOSTS:"
	CR_PROBE_SLAVE_CONNECT         Errno = 2024
	CR_PROBE_SLAVE_CONNECT_STR     Errstr ="Error connecting to slave:"
	CR_PROBE_MASTER_CONNECT        Errno = 2025
	CR_PROBE_MASTER_CONNECT_STR    Errstr ="Error connecting to master:"
	CR_SSL_CONNECTION_ERROR        Errno = 2026
	CR_SSL_CONNECTION_ERROR_STR    Errstr ="SSL connection error"
	CR_MALFORMED_PACKET            Errno = 2027
	CR_MALFORMED_PACKET_STR        Errstr ="Malformed Packet"
	CR_WRONG_LICENSE               Errno = 2028
	CR_WRONG_LICENSE_STR           Errstr ="This client library is licensed only for use with MySQL servers having '%s' license"
	CR_NULL_POINTER                Errno = 2029
	CR_NULL_POINTER_STR            Errstr ="Invalid use of null pointer"
	CR_NO_PREPARE_STMT             Errno = 2030
	CR_NO_PREPARE_STMT_STR         Errstr ="Statement not prepared"
	CR_PARAMS_NOT_BOUND            Errno = 2031
	CR_PARAMS_NOT_BOUND_STR        Errstr ="No data supplied for parameters in prepared statement"
	CR_DATA_TRUNCATED              Errno = 2032
	CR_DATA_TRUNCATED_STR          Errstr ="Data truncated"
	CR_NO_PARAMETERS_EXISTS        Errno = 2033
	CR_NO_PARAMETERS_EXISTS_STR    Errstr ="No parameters exist in the statement"
	CR_INVALID_PARAMETER_NO        Errno = 2034
	CR_INVALID_PARAMETER_NO_STR    Errstr ="Invalid parameter number"
	CR_INVALID_BUFFER_USE          Errno = 2035
	CR_INVALID_BUFFER_USE_STR      Errstr ="Can't send long data for non-string/non-binary data types (parameter: %d)"
	CR_UNSUPPORTED_PARAM_TYPE      Errno = 2036
	CR_UNSUPPORTED_PARAM_TYPE_STR  Errstr ="Using unsupported parameter type: %s (parameter: %d)"
	CR_CONN_UNKNOW_PROTOCOL        Errno = 2047
	CR_CONN_UNKNOW_PROTOCOL_STR    Errstr ="Wrong or unknown protocol"
	CR_SECURE_AUTH                 Errno = 2049
	CR_SECURE_AUTH_STR             Errstr ="Connection using old (pre-4.1.1) authentication protocol refused (client option 'secure_auth' enabled)"
	CR_FETCH_CANCELED              Errno = 2050
	CR_FETCH_CANCELED_STR          Errstr ="Row retrieval was canceled by mysql_stmt_close() call"
	CR_NO_DATA                     Errno = 2051
	CR_NO_DATA_STR                 Errstr ="Attempt to read column without prior row fetch"
	CR_NO_STMT_METADATA            Errno = 2052
	CR_NO_STMT_METADATA_STR        Errstr ="Prepared statement contains no metadata"
	CR_NO_RESULT_SET               Errno = 2053
	CR_NO_RESULT_SET_STR           Errstr ="Attempt to read a row while there is no result set associated with the statement"
	CR_NOT_IMPLEMENTED             Errno = 2054
	CR_NOT_IMPLEMENTED_STR         Errstr ="This feature is not implemented yet"
	CR_SERVER_LOST_EXTENDED        Errno = 2055
	CR_SERVER_LOST_EXTENDED_STR    Errstr ="Lost connection to MySQL server at '%s', system error: %d"
	CR_STMT_CLOSED                 Errno = 2056
	CR_STMT_CLOSED_STR             Errstr ="Statement closed indirectly because of a preceeding %s() call"
	CR_NEW_STMT_METADATA           Errno = 2057
	CR_NEW_STMT_METADATA_STR       Errstr ="The number of columns in the result set differs from the number of bound buffers. You must reset the statement, rebind the result set columns, and execute the statement again"
	CR_ALREADY_CONNECTED           Errno = 2058
	CR_ALREADY_CONNECTED_STR       Errstr ="This handle is already connected"
	CR_AUTH_PLUGIN_CANNOT_LOAD     Errno = 2059
	CR_AUTH_PLUGIN_CANNOT_LOAD_STR Errstr ="Authentication plugin '%s' cannot be loaded: %s"
)

// Client error struct
type ClientError struct {
	Errno Errno
	Errstr Errstr
}

// Convert to string
func (e *ClientError) Error() string {
	return fmt.Sprintf("#%d %s", e.Errno, e.Errstr)
}

// Server error struct
type ServerError struct {
	Errno Errno
	Errstr Errstr
}

// Convert to string
func (e *ServerError) Error() string {
	return fmt.Sprintf("#%d %s", e.Errno, e.Errstr)
}
