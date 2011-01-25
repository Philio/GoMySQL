// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

type commands byte

const (
	COM_QUIT                commands = 0x01
	COM_INIT_DB             commands = 0x02
	COM_QUERY               commands = 0x03
	COM_FIELD_LIST          commands = 0x04
	COM_CREATE_DB           commands = 0x05
	COM_DROP_DB             commands = 0x06
	COM_REFRESH             commands = 0x07
	COM_SHUTDOWN            commands = 0x08
	COM_STATISTICS          commands = 0x09
	COM_PROCESS_INFO        commands = 0x0a
	COM_CONNECT             commands = 0x0b
	COM_PROCESS_KILL        commands = 0x0c
	COM_DEBUG               commands = 0x0d
	COM_PING                commands = 0x0e
	COM_TIME                commands = 0x0f
	COM_DELAYED_INSERT      commands = 0x10
	COM_CHANGE_USER         commands = 0x11
	COM_BINLOG_DUMP         commands = 0x12
	COM_TABLE_DUMP          commands = 0x13
	COM_CONNECT_OUT         commands = 0x14
	COM_REGISTER_SLAVE      commands = 0x15
	COM_STMT_PREPARE        commands = 0x16
	COM_STMT_EXECUTE        commands = 0x17
	COM_STMT_SEND_LONG_DATA commands = 0x18
	COM_STMT_CLOSE          commands = 0x19
	COM_STMT_RESET          commands = 0x1a
	COM_SET_OPTION          commands = 0x1b
	COM_STMT_FETCH          commands = 0x1c
)

type ClientFlags uint32

const (
	CLIENT_LONG_PASSWORD    ClientFlags = 1 << iota
	CLIENT_FOUND_ROWS       ClientFlags = 1 << iota
	CLIENT_LONG_FLAG        ClientFlags = 1 << iota
	CLIENT_CONNECT_WITH_DB  ClientFlags = 1 << iota
	CLIENT_NO_SCHEMA        ClientFlags = 1 << iota
	CLIENT_COMPRESS         ClientFlags = 1 << iota
	CLIENT_ODBC             ClientFlags = 1 << iota
	CLIENT_LOCAL_FILES      ClientFlags = 1 << iota
	CLIENT_IGNORE_SPACE     ClientFlags = 1 << iota
	CLIENT_PROTOCOL_41      ClientFlags = 1 << iota
	CLIENT_INTERACTIVE      ClientFlags = 1 << iota
	CLIENT_SSL              ClientFlags = 1 << iota
	CLIENT_IGNORE_SIGPIPE   ClientFlags = 1 << iota
	CLIENT_TRANSACTIONS     ClientFlags = 1 << iota
	CLIENT_RESERVED         ClientFlags = 1 << iota
	CLIENT_SECURE_CONN      ClientFlags = 1 << iota
	CLIENT_MULTI_STATEMENTS ClientFlags = 1 << iota
	CLIENT_MULTI_RESULTS    ClientFlags = 1 << iota
)

type ServerStatus uint16

const (
	SERVER_STATUS_IN_TRANS             ServerStatus = 0x01
	SERVER_STATUS_AUTOCOMMIT           ServerStatus = 0x02
	SERVER_MORE_RESULTS_EXISTS         ServerStatus = 0x08
	SERVER_QUERY_NO_GOOD_INDEX_USED    ServerStatus = 0x10
	SERVER_QUERY_NO_INDEX_USED         ServerStatus = 0x20
	SERVER_STATUS_CURSOR_EXISTS        ServerStatus = 0x40
	SERVER_STATUS_LAST_ROW_SENT        ServerStatus = 0x80
	SERVER_STATUS_DB_DROPPED           ServerStatus = 0x100
	SERVER_STATUS_NO_BACKSLASH_ESCAPES ServerStatus = 0x200
	SERVER_STATUS_METADATA_CHANGED     ServerStatus = 0x400
)

type FieldTypes byte

const (
	FIELD_TYPE_DECIMAL     FieldTypes = 0x00
	FIELD_TYPE_TINY        FieldTypes = 0x01
	FIELD_TYPE_SHORT       FieldTypes = 0x02
	FIELD_TYPE_LONG        FieldTypes = 0x03
	FIELD_TYPE_FLOAT       FieldTypes = 0x04
	FIELD_TYPE_DOUBLE      FieldTypes = 0x05
	FIELD_TYPE_NULL        FieldTypes = 0x06
	FIELD_TYPE_TIMESTAMP   FieldTypes = 0x07
	FIELD_TYPE_LONGLONG    FieldTypes = 0x08
	FIELD_TYPE_INT24       FieldTypes = 0x09
	FIELD_TYPE_DATE        FieldTypes = 0x0a
	FIELD_TYPE_TIME        FieldTypes = 0x0b
	FIELD_TYPE_DATETIME    FieldTypes = 0x0c
	FIELD_TYPE_YEAR        FieldTypes = 0x0d
	FIELD_TYPE_NEWDATE     FieldTypes = 0x0e
	FIELD_TYPE_VARCHAR     FieldTypes = 0x0f
	FIELD_TYPE_BIT         FieldTypes = 0x10
	FIELD_TYPE_NEWDECIMAL  FieldTypes = 0xf6
	FIELD_TYPE_ENUM        FieldTypes = 0xf7
	FIELD_TYPE_SET         FieldTypes = 0xf8
	FIELD_TYPE_TINY_BLOB   FieldTypes = 0xf9
	FIELD_TYPE_MEDIUM_BLOB FieldTypes = 0xfa
	FIELD_TYPE_LONG_BLOB   FieldTypes = 0xfb
	FIELD_TYPE_BLOB        FieldTypes = 0xfc
	FIELD_TYPE_VAR_STRING  FieldTypes = 0xfd
	FIELD_TYPE_STRING      FieldTypes = 0xfe
	FIELD_TYPE_GEOMETRY    FieldTypes = 0xff
)

type FieldFlags uint16

const (
	FLAG_NOT_NULL       FieldFlags = 1 << iota
	FLAG_PRI_KEY        FieldFlags = 1 << iota
	FLAG_UNIQUE_KEY     FieldFlags = 1 << iota
	FLAG_MULTIPLE_KEY   FieldFlags = 1 << iota
	FLAG_BLOB           FieldFlags = 1 << iota
	FLAG_UNSIGNED       FieldFlags = 1 << iota
	FLAG_ZEROFILL       FieldFlags = 1 << iota
	FLAG_BINARY         FieldFlags = 1 << iota
	FLAG_ENUM           FieldFlags = 1 << iota
	FLAG_AUTO_INCREMENT FieldFlags = 1 << iota
	FLAG_TIMESTAMP      FieldFlags = 1 << iota
	FLAG_SET            FieldFlags = 1 << iota
	FLAG_UNKNOWN_1      FieldFlags = 1 << iota
	FLAG_UNKNOWN_2      FieldFlags = 1 << iota
	FLAG_UNKNOWN_3      FieldFlags = 1 << iota
	FLAG_UNKNOWN_4      FieldFlags = 1 << iota
)

type ExecuteFlags uint8

const (
	CURSOR_TYPE_NO_CURSOR  ExecuteFlags = 0
	CURSOR_TYPE_READ_ONLY  ExecuteFlags = 1 << iota
	CURSOR_TYPE_FOR_UPDATE ExecuteFlags = 1 << iota
	CURSOR_TYPE_SCROLLABLE ExecuteFlags = 1 << iota
)
