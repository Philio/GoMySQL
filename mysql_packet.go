/**
 * GoMySQL - A MySQL client library for Go
 * Copyright 2010 Phil Bayfield
 * This software is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License
 * Further information on this license can be found here: http://creativecommons.org/licenses/by-sa/2.0/uk/
 */
package mysql

import (
	"bufio"
	"io"
	"os"
	"crypto/sha1"
	"strconv"
	"math"
)

/**
 * The packet header, always at the start of every packet
 */
type packetHeader struct {
	packetFunctions
	length   uint32
	sequence uint8
}

/**
 * Read packet header from buffer
 */
func (hdr *packetHeader) read(reader *bufio.Reader) (err os.Error) {
	// Read length
	num, err := hdr.readNumber(reader, 3)
	if err != nil {
		return
	}
	hdr.length = uint32(num)
	// Read sequence
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	hdr.sequence = uint8(c)
	return
}

/**
 * Write packet header to buffer
 */
func (hdr *packetHeader) write(writer *bufio.Writer) (err os.Error) {
	// Write length
	err = hdr.writeNumber(writer, uint64(hdr.length), 3)
	if err != nil {
		return
	}
	// Write sequence
	err = writer.WriteByte(byte(hdr.sequence))
	if err != nil {
		return
	}
	return
}

/**
 * The initialisation packet (sent from server to client immediately following new connection
 */
type packetInit struct {
	packetFunctions
	protocolVersion uint8
	serverVersion   string
	threadId        uint32
	scrambleBuff    []byte
	serverCaps      uint16
	serverLanguage  uint8
	serverStatus    uint16
}

/**
 * Read initialisation packet from buffer
 */
func (pkt *packetInit) read(reader *bufio.Reader) (err os.Error) {
	// Need to keep track of bytes read for 5.5 compatibility
	bytesRead := uint32(0)
	// Protocol version
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.protocolVersion = uint8(c)
	bytesRead += 1
	// Server version
	line, err := reader.ReadString(0x00)
	if err != nil {
		return
	}
	pkt.serverVersion = line
	bytesRead += uint32(len(line))
	// Thread id
	num, err := pkt.readNumber(reader, 4)
	if err != nil {
		return
	}
	pkt.threadId = uint32(num)
	bytesRead += 4
	// Scramble buffer (first part)
	pkt.scrambleBuff = make([]byte, 20)
	_, err = io.ReadFull(reader, pkt.scrambleBuff[0:8])
	if err != nil {
		return
	}
	bytesRead += 8
	// Skip next byte
	err = pkt.readFill(reader, 1)
	if err != nil {
		return
	}
	bytesRead += 1
	// Server capabilities
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.serverCaps = uint16(num)
	bytesRead += 2
	// Server language
	c, err = reader.ReadByte()
	if err != nil {
		return
	}
	pkt.serverLanguage = uint8(c)
	bytesRead += 1
	// Server status
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.serverStatus = uint16(num)
	bytesRead += 2
	// Skip next 13 bytes
	err = pkt.readFill(reader, 13)
	if err != nil {
		return
	}
	bytesRead += 13
	// Scramble buffer (second part)
	_, err = io.ReadFull(reader, pkt.scrambleBuff[8:20])
	if err != nil {
		return
	}
	bytesRead += 12
	// Read final byte
	err = pkt.readFill(reader, 1)
	bytesRead += 1
	// Read additional bytes (5.5 support)
	if bytesRead < pkt.header.length {
		bytes := make([]byte, pkt.header.length-bytesRead)
		_, err = io.ReadFull(reader, bytes)
	}
	return
}

/**
 * The authentication packet (sent from client to server following the initialisation packet)
 */
type packetAuth struct {
	packetFunctions
	clientFlags   uint32
	maxPacketSize uint32
	charsetNumber uint8
	user          string
	scrambleBuff  []byte
	database      string
}

/**
 * Password encryption mechanism use by MySQL = SHA1(SHA1(SHA1(password)), scramble) XOR SHA1(password)
 */
func (pkt *packetAuth) encrypt(password string, scrambleBuff []byte) {
	// Convert password to byte array
	passbytes := []byte(password)
	// stage1_hash = SHA1(password)
	// SHA1 encode
	crypt := sha1.New()
	crypt.Write(passbytes)
	stg1Hash := crypt.Sum()
	// token = SHA1(SHA1(stage1_hash), scramble) XOR stage1_hash
	// SHA1 encode again
	crypt.Reset()
	crypt.Write(stg1Hash)
	stg2Hash := crypt.Sum()
	// SHA1 2nd hash and scramble
	crypt.Reset()
	crypt.Write(scrambleBuff)
	crypt.Write(stg2Hash)
	stg3Hash := crypt.Sum()
	// XOR with first hash
	pkt.scrambleBuff = make([]byte, 21)
	pkt.scrambleBuff[0] = 0x14
	for i := 0; i < 20; i++ {
		pkt.scrambleBuff[i+1] = stg3Hash[i] ^ stg1Hash[i]
	}
}

/**
 * Write authentication packet to buffer and flush
 */
func (pkt *packetAuth) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 33 + uint32(len(pkt.user))
	if len(pkt.scrambleBuff) > 0 {
		pkt.header.length += 21
	} else {
		pkt.header.length += 1
	}
	if len(pkt.database) > 0 {
		pkt.header.length += uint32(len(pkt.database)) + 1
	}
	pkt.header.sequence = 1
	err = pkt.header.write(writer)
	if err != nil {
		return
	}
	// Write the client flags
	err = pkt.writeNumber(writer, uint64(pkt.clientFlags), 4)
	if err != nil {
		return
	}
	// Write max packet size
	err = pkt.writeNumber(writer, uint64(pkt.maxPacketSize), 4)
	if err != nil {
		return
	}
	// Write charset
	err = writer.WriteByte(byte(pkt.charsetNumber))
	if err != nil {
		return
	}
	// Filler 23 x 0x00
	err = pkt.writeFill(writer, 23)
	if err != nil {
		return
	}
	// Write username
	_, err = writer.WriteString(pkt.user)
	if err != nil {
		return
	}
	// Terminate string with 0x00
	err = pkt.writeFill(writer, 1)
	if err != nil {
		return
	}
	// Write Scramble buffer of filler 1 x 0x00
	if len(pkt.scrambleBuff) > 0 {
		_, err = writer.Write(pkt.scrambleBuff)
	} else {
		err = pkt.writeFill(writer, 1)
	}
	if err != nil {
		return
	}
	// Write database name
	if len(pkt.database) > 0 {
		_, err = writer.WriteString(pkt.database)
		if err != nil {
			return
		}
		// Terminate string with 0x00
		err = pkt.writeFill(writer, 1)
		if err != nil {
			return
		}
	}
	// Flush
	err = writer.Flush()
	return
}

/**
 * The OK packet (received after an operation completed successfully)
 */
type packetOK struct {
	packetFunctions
	fieldCount   uint8
	affectedRows uint64
	insertId     uint64
	serverStatus uint16
	warningCount uint16
	message      string
}

/**
 * Read the OK packet from the buffer
 */
func (pkt *packetOK) read(reader *bufio.Reader) (err os.Error) {
	var n, bytes int
	// Read field count
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.fieldCount = uint8(c)
	// Read affected rows
	pkt.affectedRows, n, err = pkt.readlengthCodedBinary(reader)
	if err != nil {
		return
	}
	bytes = n
	// Read insert id
	pkt.insertId, n, err = pkt.readlengthCodedBinary(reader)
	if err != nil {
		return
	}
	bytes += n
	// Read server status
	num, err := pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.serverStatus = uint16(num)
	// Read warning count
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.warningCount = uint16(num)
	// Read message
	bytes += 5
	if int(pkt.header.length) > bytes {
		msg := make([]byte, int(pkt.header.length)-bytes)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			return
		}
		pkt.message = string(msg)
	}
	return
}

/**
 * Error packet (received after an operation failed)
 */
type packetError struct {
	packetFunctions
	fieldCount uint8
	errno      uint16
	state      string
	error      string
}

/**
 * Read error packet from the buffer
 */
func (pkt *packetError) read(reader *bufio.Reader) (err os.Error) {
	var bytes int
	// Read field count
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.fieldCount = uint8(c)
	// Read error code
	num, err := pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.errno = uint16(num)
	// Read # byte
	c, err = reader.ReadByte()
	if err != nil {
		return
	}
	// If byte isn't # then state is missing
	if c == 0x23 {
		// Read state
		state := make([]byte, 5)
		_, err = io.ReadFull(reader, state)
		if err != nil {
			return
		}
		pkt.state = string(state)
		bytes = 9
	} else {
		reader.UnreadByte()
		bytes = 3
	}
	// Read message
	if int(pkt.header.length) > bytes {
		msg := make([]byte, int(pkt.header.length)-bytes)
		_, err = io.ReadFull(reader, msg)
		if err != nil {
			return
		}
		pkt.error = string(msg)
	}
	return
}

/**
 * Command packet (send command + optional args)
 */
type packetCommand struct {
	packetFunctions
	command byte
	args    []interface{}
}

/**
 * Write the command packet to the buffer and send to the server
 */
func (pkt *packetCommand) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 1
	// Calculate packet length
	for i := 0; i < len(pkt.args); i++ {
		switch v := pkt.args[i].(type) {
		default:
			return os.ErrorString("Unsupported type")
		case string: pkt.header.length += uint32(len(v))
		case uint8: pkt.header.length += 1
		case uint16: pkt.header.length += 2
		case uint32: pkt.header.length += 4
		case uint64: pkt.header.length += 8
		}
	}
	pkt.header.sequence = 0
	err = pkt.header.write(writer)
	if err != nil {
		return
	}
	// Write command
	err = writer.WriteByte(byte(pkt.command))
	if err != nil {
		return
	}
	// Write params
	for i := 0; i < len(pkt.args); i++ {
		switch v := pkt.args[i].(type) {
		case string: _, err = writer.WriteString(v)
		case uint8: err = pkt.writeNumber(writer, uint64(v), 1)
		case uint16: err = pkt.writeNumber(writer, uint64(v), 2)
		case uint32: err = pkt.writeNumber(writer, uint64(v), 4)
		case uint64: err = pkt.writeNumber(writer, v, 8)
		}
		if err != nil {
			return
		}
	}
	// Flush
	err = writer.Flush()
	return
}

/**
 * The result set packet (used only to identify a result set)
 */
type packetResultSet struct {
	packetFunctions
	fieldCount uint64
	extra      uint64
}

/**
 * Read the result set from the buffer (requires the packet length as buffer will contain several packets)
 */
func (pkt *packetResultSet) read(reader *bufio.Reader) (err os.Error) {
	var n int
	// Read field count
	pkt.fieldCount, n, err = pkt.readlengthCodedBinary(reader)
	if err != nil {
		return
	}
	// Read extra
	if n < int(pkt.header.length) {
		pkt.extra, _, err = pkt.readlengthCodedBinary(reader)
	}
	return
}

/**
 * Field packet
 */
type packetField struct {
	packetFunctions
	catalog       string
	database      string
	table         string
	orgtable      string
	name          string
	orgName       string
	charsetNumber uint16
	length        uint32
	fieldType     byte
	flags         uint16
	decimals      uint8
	fieldDefault  uint64
}

/**
 * Read field packet from buffer
 */
func (pkt *packetField) read(reader *bufio.Reader) (err os.Error) {
	var n, bytes int
	// Read catalog
	pkt.catalog, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes = n
	// Read database name
	pkt.database, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes += n
	// Read table name
	pkt.table, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes += n
	// Read original table name
	pkt.orgtable, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes += n
	// Read name
	pkt.name, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes += n
	// Read original name
	pkt.orgName, n, err = pkt.readlengthCodedString(reader)
	if err != nil {
		return
	}
	bytes += n
	// Filler 1 x 0x00
	err = pkt.readFill(reader, 1)
	if err != nil {
		return
	}
	// Read charset
	num, err := pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.charsetNumber = uint16(num)
	// Read length
	num, err = pkt.readNumber(reader, 4)
	if err != nil {
		return
	}
	pkt.length = uint32(num)
	// Read type
	pkt.fieldType, err = reader.ReadByte()
	if err != nil {
		return
	}
	// Read flags
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.flags = uint16(num)
	// Read decimals
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.decimals = uint8(c)
	// Filler 2 x 0x00
	err = pkt.readFill(reader, 2)
	if err != nil {
		return
	}
	bytes += 13
	// Read default if set
	if int(pkt.header.length) > bytes {
		pkt.fieldDefault, _, err = pkt.readlengthCodedBinary(reader)
	}
	return
}

/**
 * Row data packet
 */
type packetRowData struct {
	packetFunctions
	fields     []*MySQLField
	nullBitMap byte
	values     []interface{}
}

/**
 * Read row data packet
 */
func (pkt *packetRowData) read(reader *bufio.Reader) (err os.Error) {
	// Read (check if exists) null bit map
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	if c >= 0x40 && c <= 0x7f {
		pkt.nullBitMap = c
	} else {
		reader.UnreadByte()
	}
	// Allocate memory
	pkt.values = make([]interface{}, len(pkt.fields))
	// Read data for each field
	for i, field := range pkt.fields {
		str, _, err := pkt.readlengthCodedString(reader)
		if err != nil {
			return
		}
		switch field.Type {
		// Strings and everythign else, keep as string
		default:
			pkt.values[i] = str
		// Tiny, small + med int convert into (u)int
		case FIELD_TYPE_TINY, FIELD_TYPE_SHORT, FIELD_TYPE_LONG:
			if field.Flags.Unsigned {
				pkt.values[i], _ = strconv.Atoui(str)
			} else {
				pkt.values[i], _ = strconv.Atoi(str)
			}
		// Big int convert to (u)int64
		case FIELD_TYPE_LONGLONG:
			if field.Flags.Unsigned {
				pkt.values[i], _ = strconv.Atoui64(str)
			} else {
				pkt.values[i], _ = strconv.Atoi64(str)
			}
		// Floats
		case FIELD_TYPE_FLOAT:
			pkt.values[i], _ = strconv.Atof32(str)
		// Double
		case FIELD_TYPE_DOUBLE:
			pkt.values[i], _ = strconv.Atof64(str)
		}
	}
	return
}

/**
 * End of file packet (received as a marker for end of packet type within a result set)
 */
type packetEOF struct {
	packetFunctions
	fieldCount   uint8
	warningCount uint16
	serverStatus uint16
}

/**
 * Read the EOF packet from the buffer
 */
func (pkt *packetEOF) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.fieldCount = uint8(c)
	// Read warning count
	num, err := pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.warningCount = uint16(num)
	// Read server status
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.serverStatus = uint16(num)
	return
}

/**
 * Ok packet for prepared statements
 */
type packetOKPrepared struct {
	packetFunctions
	fieldCount   uint8
	statementId  uint32
	columnCount  uint16
	paramCount   uint16
	warningCount uint16
}
/**
 * Read prepared statement ok packet
 */
func (pkt *packetOKPrepared) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	pkt.fieldCount = uint8(c)
	// Read statement id
	num, err := pkt.readNumber(reader, 4)
	if err != nil {
		return
	}
	pkt.statementId = uint32(num)
	// Read column count
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.columnCount = uint16(num)
	// Read param count
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.paramCount = uint16(num)
	// Read filler 1 x 0x00
	err = pkt.readFill(reader, 1)
	if err != nil {
		return
	}
	// Read warning count
	num, err = pkt.readNumber(reader, 2)
	if err != nil {
		return
	}
	pkt.warningCount = uint16(num)
	return
}

/**
 * Parameter packet
 */
type packetParameter struct {
	packetFunctions
	paramType []byte
	flags     uint16
	decimals  uint8
	length    uint32
}

/**
 * Read parameter packet
 */
func (pkt *packetParameter) read(reader *bufio.Reader) (err os.Error) {
	// Skip this packet, documentation is incorrect and it is also ignored in MySQL code!
	bytes := make([]byte, pkt.header.length)
	io.ReadFull(reader, bytes)
	return
}

/**
 * Long data packet
 */
type packetLongData struct {
	packetFunctions
	sequence    uint8
	command     byte
	statementId uint32
	paramNumber uint16
	data        string
}

/**
 * Write long data packet
 */
func (pkt *packetLongData) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 7 + uint32(len(pkt.data))
	pkt.header.sequence = pkt.sequence
	err = pkt.header.write(writer)
	if err != nil {
		return
	}
	// Write command
	err = writer.WriteByte(byte(pkt.command))
	if err != nil {
		return
	}
	// Write statement id
	err = pkt.writeNumber(writer, uint64(pkt.statementId), 4)
	if err != nil {
		return
	}
	// Write param number
	err = pkt.writeNumber(writer, uint64(pkt.paramNumber), 2)
	if err != nil {
		return
	}
	// Write data
	_, err = writer.WriteString(pkt.data)
	if err != nil {
		return
	}
	// Flush
	err = writer.Flush()
	return
}

/**
 * Execute packet
 */
type packetExecute struct {
	packetFunctions
	command        byte
	statementId    uint32
	flags          uint8
	iterationCount uint32
	nullBitMap     []byte
	newParamBound  uint8
	paramType      [][]byte
	paramData      [][]byte
	paramLength    uint32
}

/**
 * Create null bit map
 */
func (pkt *packetExecute) encodeNullBits(params []interface{}) {
	pkt.nullBitMap = make([]byte, (len(params)+7)/8)
	var bitMap uint64 = 0
	// Check if params are null (nil)
	for i := 0; i < len(params); i++ {
		if params[i] == nil {
			bitMap += 1 << uint(i)
		}
	}
	// Convert the uint64 value into bytes
	for i := 0; i < len(pkt.nullBitMap); i++ {
		pkt.nullBitMap[i] = byte(bitMap >> uint(i*8))
	}
}

/**
 * Compile param types
 */
func (pkt *packetExecute) encodeParams(params []interface{}) {
	pkt.paramType = make([][]byte, len(params))
	pkt.paramData = make([][]byte, len(params))
	// Add all param types
	for i := 0; i < len(params); i++ {
		var n uint16
		v := params[i]
		// Check for nils (NULL)
		if v == nil {
			n = uint16(FIELD_TYPE_NULL)
		} else {
			// Match go types to MySQL types
			switch value := v.(type) {
			// Strings should be length coded binary
			case string:
				n = uint16(FIELD_TYPE_STRING)
				bytes, length := pkt.packString(value)
				pkt.paramData[i] = bytes
				pkt.paramLength += uint32(length)
			// Unsigned ints simple binary encoded
			case uint:
				// uint can be 32 or 64 bits
				if strconv.IntSize == 32 {
					n = uint16(FIELD_TYPE_LONG)
					pkt.paramData[i] = pkt.packNumber(uint64(value), 4)
					pkt.paramLength += 4
				} else {
					n = uint16(FIELD_TYPE_LONGLONG)
					pkt.paramData[i] = pkt.packNumber(uint64(value), 8)
					pkt.paramLength += 8
				}
			case uint8:
				n = uint16(FIELD_TYPE_TINY)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 1)
				pkt.paramLength++
			case uint16:
				n = uint16(FIELD_TYPE_SHORT)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 2)
				pkt.paramLength += 2
			case uint32:
				n = uint16(FIELD_TYPE_LONG)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 4)
				pkt.paramLength += 4
			case uint64:
				n = uint16(FIELD_TYPE_LONGLONG)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 8)
				pkt.paramLength += 8
			// Signed ints also encoded as uint as server 'should' determine their sign based on field type
			case int:
				// int can be 32 or 64 bits
				if strconv.IntSize == 32 {
					n = uint16(FIELD_TYPE_LONG)
					pkt.paramData[i] = pkt.packNumber(uint64(value), 4)
					pkt.paramLength += 4
				} else {
					n = uint16(FIELD_TYPE_LONGLONG)
					pkt.paramData[i] = pkt.packNumber(uint64(value), 8)
					pkt.paramLength += 8
				}
			case int8:
				n = uint16(FIELD_TYPE_TINY)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 1)
				pkt.paramLength++
			case int16:
				n = uint16(FIELD_TYPE_SHORT)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 2)
				pkt.paramLength += 2
			case int32:
				n = uint16(FIELD_TYPE_LONG)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 4)
				pkt.paramLength += 4
			case int64:
				n = uint16(FIELD_TYPE_LONGLONG)
				pkt.paramData[i] = pkt.packNumber(uint64(value), 8)
				pkt.paramLength += 8
			// Floats
			case float:
				if strconv.FloatSize == 32 {
					n = uint16(FIELD_TYPE_FLOAT)
					pkt.paramData[i] = pkt.packNumber(uint64(math.Float32bits(float32(value))), 4)
					pkt.paramLength += 4
				} else {
					n = uint16(FIELD_TYPE_DOUBLE)
					pkt.paramData[i] = pkt.packNumber(uint64(math.Float64bits(float64(value))), 8)
					pkt.paramLength += 8
				}

			case float32:
				n = uint16(FIELD_TYPE_FLOAT)
				pkt.paramData[i] = pkt.packNumber(uint64(math.Float32bits(float32(value))), 4)
				pkt.paramLength += 4
			case float64:
				n = uint16(FIELD_TYPE_DOUBLE)
				pkt.paramData[i] = pkt.packNumber(uint64(math.Float64bits(float64(value))), 8)
				pkt.paramLength += 8
			}
		}
		// Add types
		pkt.paramType[i] = make([]byte, 2)
		pkt.paramType[i][0] = byte(n)
		pkt.paramType[i][1] = byte(n >> 8)
	}
}

/**
 * Write execute packet
 */
func (pkt *packetExecute) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 11 + uint32(len(pkt.nullBitMap)) + pkt.paramLength
	if pkt.newParamBound == 1 {
		pkt.header.length += uint32(len(pkt.paramType)*2)
	}
	pkt.header.sequence = 0
	err = pkt.header.write(writer)
	// Write command
	err = writer.WriteByte(byte(pkt.command))
	if err != nil {
		return
	}
	// Write statement id
	err = pkt.writeNumber(writer, uint64(pkt.statementId), 4)
	if err != nil {
		return
	}
	// Write flags
	err = writer.WriteByte(byte(pkt.flags))
	if err != nil {
		return
	}
	// Write iteration count
	err = pkt.writeNumber(writer, uint64(pkt.iterationCount), 4)
	if err != nil {
		return
	}
	// Write null bit map
	_, err = writer.Write(pkt.nullBitMap)
	if err != nil {
		return
	}
	// Write new parameter bound flag
	err = writer.WriteByte(byte(pkt.newParamBound))
	if err != nil {
		return
	}
	if pkt.newParamBound == 1 {
		// Write param types
		if len(pkt.paramType) > 0 {
			for _, paramType := range pkt.paramType {
				_, err = writer.Write(paramType)
				if err != nil {
					return
				}
			}
		}
	}
	// Write param data
	if len(pkt.paramData) > 0 {
		for _, paramData := range pkt.paramData {
			if len(paramData) > 0 {
				_, err = writer.Write(paramData)
				if err != nil {
					return
				}
			}
		}
	}
	// Flush
	err = writer.Flush()
	return
}

type packetBinaryRowData struct {
	packetFunctions
	fields []*MySQLField
	values []interface{}
}

func (pkt *packetBinaryRowData) read(reader *bufio.Reader) (err os.Error) {
	// Keep track of bytes read
	bytesRead := uint32(0)
	// Ignore first byte
	err = pkt.readFill(reader, 1)
	if err != nil {
		return
	}
	bytesRead++
	// Read null bit map
	nullBytes := (len(pkt.fields) + 9) / 8
	nullBitMap := make([]byte, nullBytes)
	_, err = io.ReadFull(reader, nullBitMap)
	if err != nil {
		return
	}
	bytesRead += uint32(nullBytes)
	// Allocate memory
	pkt.values = make([]interface{}, len(pkt.fields))
	// Read data for each field
	for i, field := range pkt.fields {
		// Check if field is null
		posByte := (i + 2) / 8
		posBit := i - (posByte * 8) + 2
		if nullBitMap[posByte]&(1<<uint8(posBit)) != 0 {
			pkt.values[i] = nil
			continue
		}
		switch field.Type {
		// Tiny int (8 bit int unsigned or signed)
		case FIELD_TYPE_TINY:
			num, err := reader.ReadByte()
			if err != nil {
				return
			}
			if field.Flags.Unsigned {
				pkt.values[i] = uint8(num)
			} else {
				pkt.values[i] = int8(num)
			}
			bytesRead++
		// Small int (16 bit int unsigned or signed)
		case FIELD_TYPE_SHORT, FIELD_TYPE_YEAR:
			num, err := pkt.readNumber(reader, 2)
			if err != nil {
				return
			}
			if field.Flags.Unsigned {
				pkt.values[i] = uint16(num)
			} else {
				pkt.values[i] = int16(num)
			}
			bytesRead += 2
		// Int (32 bit int unsigned or signed) and medium int which is actually in int32 format
		case FIELD_TYPE_LONG, FIELD_TYPE_INT24:
			num, err := pkt.readNumber(reader, 4)
			if err != nil {
				return
			}
			if field.Flags.Unsigned {
				pkt.values[i] = uint32(num)
			} else {
				pkt.values[i] = int32(num)
			}
			bytesRead += 4
		// Big int (64 bit int unsigned or signed)
		case FIELD_TYPE_LONGLONG:
			num, err := pkt.readNumber(reader, 8)
			if err != nil {
				return
			}
			if field.Flags.Unsigned {
				pkt.values[i] = num
			} else {
				pkt.values[i] = int64(num)
			}
			bytesRead += 8
		// Floats (Single precision floating point, 32 bit signed)
		case FIELD_TYPE_FLOAT:
			num, err := pkt.readNumber(reader, 4)
			if err != nil {
				return
			}
			pkt.values[i] = math.Float32frombits(uint32(num))
			bytesRead += 4
		// Double (Double precision floating point, 64 bit signed)
		case FIELD_TYPE_DOUBLE:
			num, err := pkt.readNumber(reader, 8)
			if err != nil {
				return
			}
			pkt.values[i] = math.Float64frombits(num)
			bytesRead += 8
		// Bit
		case FIELD_TYPE_BIT:
			c, err := reader.ReadByte()
			if err != nil {
				return
			}
			bytes := make([]byte, c)
			_, err = io.ReadFull(reader, bytes)
			pkt.values[i] = bytes
			bytesRead += uint32(c) + 1
		// Decimal and all strings, all length coded binary strings
		case FIELD_TYPE_DECIMAL, FIELD_TYPE_NEWDECIMAL, FIELD_TYPE_VARCHAR, FIELD_TYPE_TINY_BLOB, FIELD_TYPE_MEDIUM_BLOB, FIELD_TYPE_LONG_BLOB,
			FIELD_TYPE_BLOB, FIELD_TYPE_VAR_STRING, FIELD_TYPE_STRING:
			str, n, err := pkt.readlengthCodedString(reader)
			if err != nil {
				return
			}
			pkt.values[i] = str
			bytesRead += uint32(n)
		// Date/Datetime/Timestamp YYYY-MM-DD HH:MM:SS (From libmysql/libmysql.c read_binary_datetime)
		case FIELD_TYPE_DATE, FIELD_TYPE_TIMESTAMP, FIELD_TYPE_DATETIME:
			num, n, err := pkt.readlengthCodedBinary(reader)
			// Check if 0 bytes, just return 0 date/time format
			if num == 0 {
				if field.Type == FIELD_TYPE_DATE {
					pkt.values[i] = "0000-00-00"
				} else {
					pkt.values[i] = "0000-00-00 00:00:00"
				}
				bytesRead += uint32(n)
				break
			}
			if err != nil {
				return
			}
			// Year
			year, err := pkt.readNumber(reader, 2)
			dateStr := strconv.Uitoa64(year) + "-"
			if err != nil {
				return
			}
			// Month
			c, err := reader.ReadByte()
			if err != nil {
				return
			}
			month := uint64(c)
			if month < 10 {
				dateStr += "0"
			}
			dateStr += strconv.Uitoa64(month) + "-"
			// Day
			c, err = reader.ReadByte()
			if err != nil {
				return
			}
			day := uint64(c)
			if day < 10 {
				dateStr += "0"
			}
			dateStr += strconv.Uitoa64(day)
			if num > 4 {
				dateStr += " "
				// Hour
				c, err = reader.ReadByte()
				if err != nil {
					return
				}
				hour := uint64(c)
				if hour < 10 {
					dateStr += "0"
				}
				dateStr += strconv.Uitoa64(hour) + ":"
				// Minute
				c, err = reader.ReadByte()
				if err != nil {
					return
				}
				mins := uint64(c)
				if mins < 10 {
					dateStr += "0"
				}
				dateStr += strconv.Uitoa64(mins) + ":"
				// Seconds
				c, err = reader.ReadByte()
				if err != nil {
					return
				}
				secs := uint64(c)
				if secs < 10 {
					dateStr += "0"
				}
				dateStr += strconv.Uitoa64(secs)
			}
			pkt.values[i] = dateStr
			if num > 7 {
				err = pkt.readFill(reader, int(num-7))
				if err != nil {
					return
				}
			}
			bytesRead += uint32(num) + uint32(n)
		// Time  (From libmysql/libmysql.c read_binary_time)
		case FIELD_TYPE_TIME:
			var dateStr string
			num, n, err := pkt.readlengthCodedBinary(reader)
			// Check if 0 bytes, just return 0 time format
			if num == 0 {
				pkt.values[i] = "00:00:00"
				bytesRead += uint32(n)
				break
			}
			if err != nil {
				return
			}
			// Ignore first byte, corresponds to tm->neg in libmysql.c, unknown usage
			err = pkt.readFill(reader, 1)
			if err != nil {
				return
			}
			// Read day
			day, err := pkt.readNumber(reader, 4)
			if err != nil {
				return
			}
			// Hour
			c, err := reader.ReadByte()
			if err != nil {
				return
			}
			hour := uint64(c)
			hour += day * 24
			if hour < 10 {
				dateStr += "0"
			}
			dateStr += strconv.Uitoa64(hour) + ":"
			// Minute
			c, err = reader.ReadByte()
			if err != nil {
				return
			}
			mins := uint64(c)
			if mins < 10 {
				dateStr += "0"
			}
			dateStr += strconv.Uitoa64(mins) + ":"
			// Seconds
			c, err = reader.ReadByte()
			if err != nil {
				return
			}
			secs := uint64(c)
			if secs < 10 {
				dateStr += "0"
			}
			dateStr += strconv.Uitoa64(secs)
			pkt.values[i] = dateStr
			if num > 8 {
				err = pkt.readFill(reader, int(num-8))
				if err != nil {
					return
				}
			}
			bytesRead += uint32(num) + uint32(n)
		// Geometry types, get array of bytes
		case FIELD_TYPE_GEOMETRY:
			c, _, err := pkt.readlengthCodedBinary(reader)
			if err != nil {
				return
			}
			bytes := make([]byte, c)
			_, err = io.ReadFull(reader, bytes)
			pkt.values[i] = bytes
			bytesRead += uint32(c) + 1
		}
	}
	// In some circumstances packets seam to contain extra data, if not all
	// data has been read, read it and discard it. In reality this will
	// probably never happen unless using some very odd queries!
	if bytesRead < pkt.header.length {
		bytes := make([]byte, pkt.header.length-bytesRead)
		_, err = io.ReadFull(reader, bytes)
	}
	return
}

/**
 * Generic packet fucntions, used by all/most packets
 */
type packetFunctions struct {
	header *packetHeader
}

/**
 * Read a number from the buffer that is n bytes long
 */
func (pkt *packetFunctions) readNumber(reader *bufio.Reader, n uint8) (num uint64, err os.Error) {
	p := make([]byte, n)
	_, err = io.ReadFull(reader, p)
	if err != nil {
		return
	}
	num = 0
	for i := uint8(0); i < n; i++ {
		num |= uint64(p[i]) << (i * 8)
	}
	return
}

/**
 * Read n 0x00 bytes from the buffer
 */
func (pkt *packetFunctions) readFill(reader *bufio.Reader, n int) (err os.Error) {
	p := make([]byte, n)
	_, err = io.ReadFull(reader, p)
	return
}

/**
 * Read a length coded bunary number from the buffer
 */
func (pkt *packetFunctions) readlengthCodedBinary(reader *bufio.Reader) (num uint64, n int, err os.Error) {
	// Read first byte
	c, err := reader.ReadByte()
	if err != nil {
		return
	}
	// Determine data type
	switch {
	// 0-250 = value of first byte
	case uint8(c) <= 250:
		num = uint64(c)
		n = 1
	// 251 column value = NULL
	case uint8(c) == 251:
		num = uint64(0)
		n = 1
	// 252 following 2 = value of following 16-bit word
	case uint8(c) == 252:
		num, err = pkt.readNumber(reader, 2)
		n = 3
	// 253 following 3 = value of following 24-bit word
	case uint8(c) == 253:
		num, err = pkt.readNumber(reader, 3)
		n = 4
	// 254 following 8 = value of following 64-bit word
	case uint8(c) == 254:
		num, err = pkt.readNumber(reader, 8)
		n = 9
	}
	return
}

/**
 * Read a length coded binary string from the buffer
 */
func (pkt *packetFunctions) readlengthCodedString(reader *bufio.Reader) (str string, n int, err os.Error) {
	// Get string length
	strlen, n, err := pkt.readlengthCodedBinary(reader)
	if err != nil {
		return
	}
	// Total bytes read = n + strlen
	n += int(strlen)
	// Read string
	p := make([]byte, strlen)
	_, err = io.ReadFull(reader, p)
	if err != nil {
		return
	}
	str = string(p)
	return
}

/**
 * Write a number to the buffer using n bytes
 */
func (pkt *packetFunctions) writeNumber(writer *bufio.Writer, num uint64, n uint8) (err os.Error) {
	p := make([]byte, n)
	for i := uint8(0); i < n; i++ {
		p[i] = byte(num >> (i * 8))
	}
	_, err = writer.Write(p)
	return
}

/**
 * Write n 0x00 bytes to the buffer
 */
func (pkt *packetFunctions) writeFill(writer *bufio.Writer, n int) (err os.Error) {
	p := make([]byte, n)
	_, err = writer.Write(p)
	return
}

/**
 * Pack an unsigned int into binary format
 */
func (pkt *packetFunctions) packNumber(num uint64, n uint8) (p []byte) {
	p = make([]byte, n)
	for i := uint8(0); i < n; i++ {
		p[i] = byte(num >> (i * 8))
	}
	return
}

/**
 * Pack an string into length coded binary format
 */
func (pkt *packetFunctions) packString(str string) (p []byte, n int) {
	switch {
	// <= 250 = 1 byte
	case len(str) <= 250:
		p = make([]byte, len(str)+1)
		p[0] = byte(len(str))
		n = 1
	// <= 0xffff = 252 + 2 bytes
	case len(str) <= 0xffff:
		p = make([]byte, len(str)+3)
		p[0] = byte(252)
		p[1] = byte(len(str))
		p[2] = byte(len(str) >> 8)
		n = 3
	// <= 0xffffff = 253 + 3 bytes
	case len(str) <= 0xffffff:
		p = make([]byte, len(str)+4)
		p[0] = byte(253)
		p[1] = byte(len(str))
		p[2] = byte(len(str) >> 8)
		p[3] = byte(len(str) >> 16)
		n = 4
	}
	// Convert string to bytes
	bytes := []byte(str)
	for i := 0; i < len(str); i++ {
		p[i+n] = bytes[i]
	}
	n += len(str)
	return
}
