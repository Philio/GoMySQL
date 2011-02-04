// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

// Result struct
type Result struct {
	// Fields
	FieldCount uint64
	fieldPos   uint64
	fields     []*Field
	
	// Rows
	rows []*Row
	
	// Storage
	mode   byte
	stored bool
}

// Field struct
type Field struct {
	Database      string
	Table         string
	Name          string
	Length        uint32
	Type          uint8
	Flags         uint16
	Decimals      uint8
}

// Row struct
type Row struct {
}
