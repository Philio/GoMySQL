/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

type ErrorNo int

const (
	CR_UNKNOWN_ERROR        = 2000
	CR_CONNECTION_ERROR     = 2002
	CR_CONN_HOST_ERROR      = 2003
	CR_SERVER_GONE_ERROR    = 2006
	CR_SERVER_HANDSHAKE_ERR = 2012
	CR_SERVER_LOST          = 2013
	CR_COMMANDS_OUT_OF_SYNC = 2014
	CR_MALFORMED_PACKET     = 2027
	CR_NO_PREPARE_STMT      = 2030
	CR_PARAMS_NOT_BOUND     = 2031
)

type ErrorStr string

const (
	CR_UNKNOWN_ERROR_STR        = "Unknown MySQL error"
	CR_CONNECTION_ERROR_STR     = "Can't connect to local MySQL server through socket '%s'"
	CR_CONN_HOST_ERROR_STR      = "Can't connect to MySQL server on '%s' (%d)"
	CR_SERVER_GONE_ERROR_STR    = "MySQL server has gone away"
	CR_SERVER_HANDSHAKE_ERR_STR = "Error in server handshake"
	CR_SERVER_LOST_STR          = "Lost connection to MySQL server during query"
	CR_COMMANDS_OUT_OF_SYNC_STR = "Commands out of sync; you can't run this command now"
	CR_MALFORMED_PACKET_STR     = "Malformed Packet"
	CR_NO_PREPARE_STMT_STR      = "Statement not prepared"
	CR_PARAMS_NOT_BOUND_STR     = "No data supplied for parameters in prepared statement"
)
