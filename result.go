// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import "os"

// Result struct
type Result struct {
	// Pointer to the client
	c *Client

	// Fields
	fieldCount uint64
	fieldPos   uint64
	fields     []*Field

	// Rows
	rowPos uint64
	rows   []Row

	// Storage
	mode    byte
	allRead bool
}

// Field type
type Field struct {
	Database string
	Table    string
	Name     string
	Length   uint32
	Type     uint8
	Flags    uint16
	Decimals uint8
}

// Row types
type Row []interface{}
type Map map[string]interface{}

// Get field count
func (r *Result) FieldCount() uint64 {
	return r.fieldCount
}

// Fetch all fields
func (r *Result) FetchFields() []*Field {
	return r.fields
}

// Get row count
func (r *Result) RowCount() uint64 {
	// Stored mode
	if r.mode == RESULT_STORED {
		return uint64(len(r.rows))
	}
	return 0
}

// Fetch a row
func (r *Result) FetchRow() Row {
	// Stored result
	if r.mode == RESULT_STORED {
		// Check if all rows have been fetched
		if r.rowPos < uint64(len(r.rows)) {
			// Increment position and return current row
			r.rowPos++
			return r.rows[r.rowPos-1]
		}
	}
	// Used result
	if r.mode == RESULT_USED {
		if r.allRead == false {
			eof, err := r.c.getRow()
			if err != nil {
				return nil
			}
			if eof {
				r.allRead = true
			} else {
				return r.rows[0]
			}
		}
	}
	return nil
}

// Fetch a map
func (r *Result) FetchMap() Map {
	// Fetch row
	row := r.FetchRow()
	if row != nil {
		rowMap := make(Map)
		for key, val := range row {
			rowMap[r.fields[key].Name] = val
		}
		return rowMap
	}
	return nil
}

// Free the result
func (r *Result) Free() (err os.Error) {
	err = r.c.FreeResult()
	return
}
