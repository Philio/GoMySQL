// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

type command byte

const (
        COM_QUIT command = iota + 1
        COM_INIT_DB
        COM_QUERY
        COM_FIELD_LIST
        COM_CREATE_DB
        COM_DROP_DB
        COM_REFRESH
        COM_SHUTDOWN
        COM_STATISTICS
        COM_PROCESS_INFO
        COM_CONNECT
        COM_PROCESS_KILL
        COM_DEBUG
        COM_PING
        COM_TIME
        COM_DELAYED_INSERT
        COM_CHANGE_USER
        COM_BINLOG_DUMP
        COM_TABLE_DUMP
        COM_CONNECT_OUT
        COM_REGISTER_SLAVE
        COM_STMT_PREPARE
        COM_STMT_EXECUTE
        COM_STMT_SEND_LONG_DATA
        COM_STMT_CLOSE
        COM_STMT_RESET
        COM_SET_OPTION
        COM_STMT_FETCH
)

type ClientFlag uint32

const (
        CLIENT_LONG_PASSWORD ClientFlag = 1 << iota
        CLIENT_FOUND_ROWS
        CLIENT_LONG_FLAG
        CLIENT_CONNECT_WITH_DB
        CLIENT_NO_SCHEMA
        CLIENT_COMPRESS
        CLIENT_ODBC
        CLIENT_LOCAL_FILES
        CLIENT_IGNORE_SPACE
        CLIENT_PROTOCOL_41
        CLIENT_INTERACTIVE
        CLIENT_SSL
        CLIENT_IGNORE_SIGPIPE
        CLIENT_TRANSACTIONS
        CLIENT_RESERVED
        CLIENT_SECURE_CONN
        CLIENT_MULTI_STATEMENTS
        CLIENT_MULTI_RESULTS
)

type ServerStatus uint16

const (
        SERVER_STATUS_IN_TRANS ServerStatus = 1 << iota
        SERVER_STATUS_AUTOCOMMIT
)

const (
        SERVER_MORE_RESULTS_EXISTS ServerStatus = 1 << (iota + 3)
        SERVER_QUERY_NO_GOOD_INDEX_USED
        SERVER_QUERY_NO_INDEX_USED
        SERVER_STATUS_CURSOR_EXISTS
        SERVER_STATUS_LAST_ROW_SENT
        SERVER_STATUS_DB_DROPPED
        SERVER_STATUS_NO_BACKSLASH_ESCAPES
        SERVER_STATUS_METADATA_CHANGED
)

type FieldType byte

const (
        FIELD_TYPE_DECIMAL FieldType = iota
        FIELD_TYPE_TINY
        FIELD_TYPE_SHORT
        FIELD_TYPE_LONG
        FIELD_TYPE_FLOAT
        FIELD_TYPE_DOUBLE
        FIELD_TYPE_NULL
        FIELD_TYPE_TIMESTAMP
        FIELD_TYPE_LONGLONG
        FIELD_TYPE_INT24
        FIELD_TYPE_DATE
        FIELD_TYPE_TIME
        FIELD_TYPE_DATETIME
        FIELD_TYPE_YEAR
        FIELD_TYPE_NEWDATE
        FIELD_TYPE_VARCHAR
        FIELD_TYPE_BIT
)

const (
        FIELD_TYPE_NEWDECIMAL FieldType = iota + 0xf6
        FIELD_TYPE_ENUM
        FIELD_TYPE_SET
        FIELD_TYPE_TINY_BLOB
        FIELD_TYPE_MEDIUM_BLOB
        FIELD_TYPE_LONG_BLOB
        FIELD_TYPE_BLOB
        FIELD_TYPE_VAR_STRING
        FIELD_TYPE_STRING
        FIELD_TYPE_GEOMETRY
)

type FieldFlag uint16

const (
        FLAG_NOT_NULL FieldFlag = 1 << iota
        FLAG_PRI_KEY
        FLAG_UNIQUE_KEY
        FLAG_MULTIPLE_KEY
        FLAG_BLOB
        FLAG_UNSIGNED
        FLAG_ZEROFILL
        FLAG_BINARY
        FLAG_ENUM
        FLAG_AUTO_INCREMENT
        FLAG_TIMESTAMP
        FLAG_SET
        FLAG_UNKNOWN_1
        FLAG_UNKNOWN_2
        FLAG_UNKNOWN_3
        FLAG_UNKNOWN_4
)

type ExecuteFlag byte

const (
        CURSOR_TYPE_NO_CURSOR ExecuteFlag = 0
        CURSOR_TYPE_READ_ONLY ExecuteFlag = 1 << iota
        CURSOR_TYPE_FOR_UPDATE
        CURSOR_TYPE_SCROLLABLE
)

type Refresh byte

const (
        REFRESH_GRANT Refresh = 1 << iota
        REFRESH_LOG
        REFRESH_TABLES
        REFRESH_HOSTS
        REFRESH_STATUS
        REFRESH_THREADS
        REFRESH_SLAVE
        REFRESH_MASTER
)

type Shutdown byte

const (
        SHUTDOWN_DEFAULT Shutdown = iota
        SHUTDOWN_WAIT_CONNECTIONS
        SHUTDOWN_WAIT_TRANSACTIONS
        SHUTDOWN_WAIT_UPDATES          Shutdown = 0x08
        SHUTDOWN_WAIT_ALL_BUFFERS      Shutdown = 0x10
        SHUTDOWN_WAIT_CRITICAL_BUFFERS Shutdown = 0x11
        KILL_QUERY                     Shutdown = 0xfe
        KILL_CONNECTION                Shutdown = 0xff
)
