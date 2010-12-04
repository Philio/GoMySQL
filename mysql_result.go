/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

/**
 * All results stored using MySQLResult
 */
type MySQLResult struct {
	AffectedRows uint64
	InsertId     uint64
	WarningCount uint16
	Message      string

	Fields     []*MySQLField
	FieldCount uint64
	fieldsRead uint64
	fieldsEOF  bool

	Rows     []*MySQLRow
	RowCount uint64
	rowsEOF  bool

	pointer int
}

/**
 * Fetch the current row (as an array)
 */
func (res *MySQLResult) FetchRow() []interface{} {
	if res.RowCount > 0 {
		if len(res.Rows) > res.pointer {
			row := res.Rows[res.pointer].Data
			res.pointer++
			return row
		}
	}
	return nil
}

/**
 * Fetch a map of the current row
 */
func (res *MySQLResult) FetchMap() map[string]interface{} {
	if res.RowCount > 0 {
		if len(res.Rows) > res.pointer {
			row := res.Rows[res.pointer].Data
			rowMap := make(map[string]interface{})
			for key, val := range row {
				rowMap[res.Fields[key].Name] = val
			}
			res.pointer++
			return rowMap
		}
	}
	return nil
}

/**
 * Reset pointer
 */
func (res *MySQLResult) Reset() {
	res.pointer = 0
}

/**
 * Field definition
 */
type MySQLField struct {
	Name     string
	Length   uint32
	Type     byte
	Flags    *MySQLFieldFlags
	Decimals uint8
}

/**
 * Field flags
 */
type MySQLFieldFlags struct {
	NotNull       bool
	PrimaryKey    bool
	UniqueKey     bool
	MultiKey      bool
	Blob          bool
	Unsigned      bool
	Zerofill      bool
	Binary        bool
	Enum          bool
	AutoIncrement bool
	Timestamp     bool
	Set           bool
}

/**
 * Process flags setting found flags as boolean true
 * @todo This would probably faster using binary
 */
func (field *MySQLFieldFlags) process(flags uint16) {
	// Check not null
	if flags&FLAG_NOT_NULL != 0 {
		field.NotNull = true
	}
	// Check pri key
	if flags&FLAG_PRI_KEY != 0 {
		field.PrimaryKey = true
	}
	// Check unique
	if flags&FLAG_UNIQUE_KEY != 0 {
		field.UniqueKey = true
	}
	// Check multi key
	if flags&FLAG_MULTIPLE_KEY != 0 {
		field.MultiKey = true
	}
	// Check blob
	if flags&FLAG_BLOB != 0 {
		field.Blob = true
	}
	// Check unsigned
	if flags&FLAG_UNSIGNED != 0 {
		field.Unsigned = true
	}
	// Check zerofill
	if flags&FLAG_ZEROFILL != 0 {
		field.Zerofill = true
	}
	// Check binary
	if flags&FLAG_BINARY != 0 {
		field.Binary = true
	}
	// Check enum
	if flags&FLAG_ENUM != 0 {
		field.Enum = true
	}
	// Check auto increment
	if flags&FLAG_AUTO_INCREMENT != 0 {
		field.AutoIncrement = true
	}
	// Check timestamp
	if flags&FLAG_TIMESTAMP != 0 {
		field.Timestamp = true
	}
	// Check flag set
	if flags&FLAG_SET != 0 {
		field.Set = true
	}
}

/**
 * Row definition
 */
type MySQLRow struct {
	Data []interface{}
}
