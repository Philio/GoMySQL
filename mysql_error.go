package mysql

type ErrorNo int

const (
	CR_UNKNOWN_ERROR	= 2000
	CR_CONNECTION_ERROR	= 2002
	CR_CONN_HOST_ERROR	= 2003
	CR_SERVER_GONE_ERROR	= 2006
	CR_SERVER_HANDSHAKE_ERR = 2012
	CR_SERVER_LOST		= 2013
	CR_COMMANDS_OUT_OF_SYNC = 2014
	CR_MALFORMED_PACKET	= 2027
)

type ErrorStr string

const (
	CR_UNKNOWN_ERROR_STR		= "Unknown MySQL error"
	CR_CONNECTION_ERROR_STR		= "Can't connect to local MySQL server through socket '%s'"
	CR_CONN_HOST_ERROR_STR		= "Can't connect to MySQL server on '%s' (%d)"
	CR_SERVER_GONE_ERROR_STR	= "MySQL server has gone away"
	CR_SERVER_HANDSHAKE_ERR_STR	= "Error in server handshake"
	CR_SERVER_LOST_STR		= "Lost connection to MySQL server during query"
	CR_COMMANDS_OUT_OF_SYNC_STR	= "Commands out of sync; you can't run this command now"
	CR_MALFORMED_PACKET_STR		= "Malformed Packet"
)
