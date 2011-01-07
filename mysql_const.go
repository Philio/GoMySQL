/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

type ClientFlags uint32

const (
	CLIENT_LONG_PASSWORD    = 1 << iota
	CLIENT_FOUND_ROWS       = 1 << iota
	CLIENT_LONG_FLAG        = 1 << iota
	CLIENT_CONNECT_WITH_DB  = 1 << iota
	CLIENT_NO_SCHEMA        = 1 << iota
	CLIENT_COMPRESS         = 1 << iota
	CLIENT_ODBC             = 1 << iota
	CLIENT_LOCAL_FILES      = 1 << iota
	CLIENT_IGNORE_SPACE     = 1 << iota
	CLIENT_PROTOCOL_41      = 1 << iota
	CLIENT_INTERACTIVE      = 1 << iota
	CLIENT_SSL              = 1 << iota
	CLIENT_IGNORE_SIGPIPE   = 1 << iota
	CLIENT_TRANSACTIONS     = 1 << iota
	CLIENT_RESERVED         = 1 << iota
	CLIENT_SECURE_CONN      = 1 << iota
	CLIENT_MULTI_STATEMENTS = 1 << iota
	CLIENT_MULTI_RESULTS    = 1 << iota
)

type Commands byte

const (
	COM_QUIT                = 0x01
	COM_INIT_DB             = 0x02
	COM_QUERY               = 0x03
	COM_FIELD_LIST          = 0x04
	COM_CREATE_DB           = 0x05
	COM_DROP_DB             = 0x06
	COM_REFRESH             = 0x07
	COM_SHUTDOWN            = 0x08
	COM_STATISTICS          = 0x09
	COM_PROCESS_INFO        = 0x0a
	COM_CONNECT             = 0x0b
	COM_PROCESS_KILL        = 0x0c
	COM_DEBUG               = 0x0d
	COM_PING                = 0x0e
	COM_TIME                = 0x0f
	COM_DELAYED_INSERT      = 0x10
	COM_CHANGE_USER         = 0x11
	COM_BINLOG_DUMP         = 0x12
	COM_TABLE_DUMP          = 0x13
	COM_CONNECT_OUT         = 0x14
	COM_REGISTER_SLAVE      = 0x15
	COM_STMT_PREPARE        = 0x16
	COM_STMT_EXECUTE        = 0x17
	COM_STMT_SEND_LONG_DATA = 0x18
	COM_STMT_CLOSE          = 0x19
	COM_STMT_RESET          = 0x1a
	COM_SET_OPTION          = 0x1b
	COM_STMT_FETCH          = 0x1c
)

type ResultPacket byte

const (
	ResultPacketOK    = 0x00
	ResultPacketError = 0xff
	ResultPacketEOF   = 0xfe
)

type FieldTypes byte

const (
	FIELD_TYPE_DECIMAL     = 0x00 // Done
	FIELD_TYPE_TINY        = 0x01 // Done
	FIELD_TYPE_SHORT       = 0x02 // Done
	FIELD_TYPE_LONG        = 0x03 // Done
	FIELD_TYPE_FLOAT       = 0x04 // Done
	FIELD_TYPE_DOUBLE      = 0x05 // Done
	FIELD_TYPE_NULL        = 0x06 // Via NULL bit map
	FIELD_TYPE_TIMESTAMP   = 0x07 // Done
	FIELD_TYPE_LONGLONG    = 0x08 // Done
	FIELD_TYPE_INT24       = 0x09 // Done
	FIELD_TYPE_DATE        = 0x0a // Done
	FIELD_TYPE_TIME        = 0x0b // Done
	FIELD_TYPE_DATETIME    = 0x0c // Done
	FIELD_TYPE_YEAR        = 0x0d // Done
	FIELD_TYPE_NEWDATE     = 0x0e // Appears not to be in use (yet?)
	FIELD_TYPE_VARCHAR     = 0x0f // Done
	FIELD_TYPE_BIT         = 0x10 // Done
	FIELD_TYPE_NEWDECIMAL  = 0xf6 // Done
	FIELD_TYPE_ENUM        = 0xf7 // Enums are sent as strings (0xfe)
	FIELD_TYPE_SET         = 0xf8 // Sets are sent as strings (0xfe)
	FIELD_TYPE_TINY_BLOB   = 0xf9 // Done
	FIELD_TYPE_MEDIUM_BLOB = 0xfa // Done
	FIELD_TYPE_LONG_BLOB   = 0xfb // Done
	FIELD_TYPE_BLOB        = 0xfc // Done
	FIELD_TYPE_VAR_STRING  = 0xfd // Done
	FIELD_TYPE_STRING      = 0xfe // Done
	FIELD_TYPE_GEOMETRY    = 0xff // Done, as byte array
)

type FieldAttribs uint16

const (
	FLAG_NOT_NULL       = 1 << iota
	FLAG_PRI_KEY        = 1 << iota
	FLAG_UNIQUE_KEY     = 1 << iota
	FLAG_MULTIPLE_KEY   = 1 << iota
	FLAG_BLOB           = 1 << iota
	FLAG_UNSIGNED       = 1 << iota
	FLAG_ZEROFILL       = 1 << iota
	FLAG_BINARY         = 1 << iota
	FLAG_ENUM           = 1 << iota
	FLAG_AUTO_INCREMENT = 1 << iota
	FLAG_TIMESTAMP      = 1 << iota
	FLAG_SET            = 1 << iota
	FLAG_UNKNOWN_1      = 1 << iota
	FLAG_UNKNOWN_2      = 1 << iota
	FLAG_UNKNOWN_3      = 1 << iota
	FLAG_UNKNOWN_4      = 1 << iota
)

type ExecuteFlags uint8

const (
	CURSOR_TYPE_NO_CURSOR  = 0
	CURSOR_TYPE_READ_ONLY  = 1 << iota
	CURSOR_TYPE_FOR_UPDATE = 1 << iota
	CURSOR_TYPE_SCROLLABLE = 1 << iota
)
