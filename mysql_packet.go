package mysql

import (
	"bufio"
	"os"
	"crypto/sha1"
)

/**
 * The packet header, always at the start of every packet
 */
type packetHeader struct {
	length		uint32
	sequence	uint8
}

/**
 * Read packet header from buffer
 */
func (hdr *packetHeader) read(reader *bufio.Reader) (err os.Error) {
	// Read length
	num, err := readNumber(reader, 3)
	if err != nil { return err }
	hdr.length = uint32(num)
	// Read sequence
	c, err := reader.ReadByte()
	if err != nil { return err }
	hdr.sequence = uint8(c)
	return
}

/**
 * Write packet header to buffer
 */
func (hdr *packetHeader) write(writer *bufio.Writer) (err os.Error) {
	// Write length
	err = writeNumber(writer, uint64(hdr.length), 3)
	if err != nil { return err }
	// Write sequence
	err = writer.WriteByte(byte(hdr.sequence))
	if err != nil { return err }
	return
}

/**
 * The initialisation packet (sent from server to client immediately following new connection
 */
type packetInit struct {
	header		*packetHeader
	protocolVersion	uint8
	serverVersion	string
	threadId	uint32
	scrambleBuff	[]byte
	serverCaps	uint16
	serverLanguage	uint8
	serverStatus	uint16
}

/**
 * Read initialisation packet from buffer
 */
func (pkt *packetInit) read(reader *bufio.Reader) (err os.Error) {
	// Protocol version
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.protocolVersion = uint8(c)
	// Server version
	line, err := reader.ReadString(0x00)
	if err != nil { return err }
	pkt.serverVersion = line
	// Thread id
	num, err := readNumber(reader, 4)
	if err != nil { return err }
	pkt.threadId = uint32(num)
	// Scramble buffer (first part)
	pkt.scrambleBuff = new([20]byte)
	_, err = reader.Read(pkt.scrambleBuff[0:8])
	if err != nil { return err }
	// Skip next byte
	err = readFill(reader, 1);
	if err != nil { return err }
	// Server capabilities
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.serverCaps = uint16(num)
	// Server language
	c, err = reader.ReadByte()
	if err != nil { return err }
	pkt.serverLanguage = uint8(c)
	// Server status
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.serverStatus = uint16(num)
	// Skip next 13 bytes
	err = readFill(reader, 13);
	if err != nil { return err }
	// Scramble buffer (second part)
	_, err = reader.Read(pkt.scrambleBuff[8:20])
	if err != nil { return err }
	// Read final byte
	err = readFill(reader, 1);
	if err != nil { return err }
	return
}

/**
 * The authentication packet (sent from client to server following the initialisation packet)
 */
type packetAuth struct {
	header		*packetHeader
	clientFlags	uint32
	maxPacketSize	uint32
	charsetNumber	uint8
	user		string
	scrambleBuff	[]byte
	database	string
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
	pkt.scrambleBuff = new([21]byte)
	pkt.scrambleBuff[0] = 0x14
	for i := 0; i < 20; i ++ {
		pkt.scrambleBuff[i + 1] = stg3Hash[i] ^ stg1Hash[i]
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
	if err != nil { return err }
	// Write the client flags
	err = writeNumber(writer, uint64(pkt.clientFlags), 4)
	if err != nil { return err }
	// Write max packet size
	err = writeNumber(writer, uint64(pkt.maxPacketSize), 4)
	if err != nil { return err }
	// Write charset
	err = writer.WriteByte(byte(pkt.charsetNumber))
	if err != nil { return err }
	// Filler 23 x 0x00
	err = writeFill(writer, 23)
	if err != nil { return err }
	// Write username
	_, err = writer.WriteString(pkt.user)
	if err != nil { return err }
	// Terminate string with 0x00
	err = writeFill(writer, 1)
	if err != nil { return err }
	// Write Scramble buffer of filler 1 x 0x00
	if len(pkt.scrambleBuff) > 0 {
		_, err = writer.Write(pkt.scrambleBuff)
	} else {
		err = writeFill(writer, 1)
	}
	if err != nil { return err }
	// Write database name
	if len(pkt.database) > 0 {
		_, err = writer.WriteString(pkt.database)
		if err != nil { return err }
		// Terminate string with 0x00
		err = writeFill(writer, 1)
		if err != nil { return err }
	}
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return
}

/**
 * The OK packet (received after an operation completed successfully)
 */
type packetOK struct {
	header		*packetHeader
	fieldCount	uint8
	affectedRows	uint64
	insertId	uint64
	serverStatus	uint16
	warningCount	uint16
	message		string
}

/**
 * Read the OK packet from the buffer
 */
func (pkt *packetOK) read(reader *bufio.Reader) (err os.Error) {
	var n, bytes int
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.fieldCount = uint8(c)
	// Read affected rows
	pkt.affectedRows, n, err = readlengthCodedBinary(reader)
	if err != nil { return err }
	bytes = n
	// Read insert id
	pkt.insertId, n, err = readlengthCodedBinary(reader)
	if err != nil { return err }
	bytes += n
	// Read server status
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.serverStatus = uint16(num)
	// Read warning count
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.warningCount = uint16(num)
	// Read message
	bytes += 5
	if int(pkt.header.length) > bytes {
		msg := make([]byte, int(pkt.header.length) - bytes)
		_, err = reader.Read(msg)
		if err != nil { return err }
		pkt.message = string(msg)
	}
	return
}

/**
 * Error packet (received after an operation failed)
 */
type packetError struct {
	header		*packetHeader
	fieldCount	uint8
	errno		uint16
	state		string
	error		string
}

/**
 * Read error packet from the buffer
 */
func (pkt *packetError) read(reader *bufio.Reader) (err os.Error) {
	var bytes int
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.fieldCount = uint8(c)
	// Read error code
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.errno = uint16(num)
	// Read # byte
	c, err = reader.ReadByte()
	if err != nil { return err }
	// If byte isn't # then state is missing
	if c == 0x23 {
		// Read state
		state := make([]byte, 5)
		_, err = reader.Read(state)
		if err != nil { return err }
		pkt.state = string(state)
		bytes = 9
	} else {
		reader.UnreadByte()
		bytes = 3
	}
	// Read message
	if int(pkt.header.length) > bytes {
		msg := make([]byte, int(pkt.header.length) - bytes)
		_, err = reader.Read(msg)
		if err != nil { return err }
		pkt.error = string(msg)
	}
	return
}

/**
 * Standard command packet (tells the server to do something defined by arg)
 */
type packetCommand struct {
	header		*packetHeader
	command		byte
	arg		string
}

/**
 * Write the command packet to the buffer and send to the server
 */
func (pkt *packetCommand) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 1 + uint32(len(pkt.arg))
	pkt.header.sequence = 0
	err = pkt.header.write(writer)
	if err != nil { return err }
	// Write command
	err = writer.WriteByte(byte(pkt.command))
	if err != nil { return err }
	// Write arg
	if len(pkt.arg) > 0 {
		_, err = writer.WriteString(pkt.arg)
		if err != nil { return err }
	}
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return
}

/**
 * The result set packet (used only to identify a result set)
 */
type packetResultSet struct {
	header		*packetHeader
	fieldCount	uint64
	extra		uint64
}

/**
 * Read the result set from the buffer (requires the packet length as buffer will contain several packets)
 */
func (pkt *packetResultSet) read(reader *bufio.Reader) (err os.Error) {
	var n int
	// Read field count
	pkt.fieldCount, n, err = readlengthCodedBinary(reader)
	if err != nil { return err }
	// Read extra
	if n < int(pkt.header.length) {
		pkt.extra, _, err = readlengthCodedBinary(reader)
		if err != nil { return err }
	}
	return nil
}

/**
 * Field packet
 */
type packetField struct {
	header		*packetHeader
	catalog		string
	database	string
	table		string
	orgtable	string
	name		string
	orgName		string
	charsetNumber	uint16
	length		uint32
	fieldType	byte
	flags		uint16
	decimals	uint8
	fieldDefault	uint64
}

/**
 * Read field packet from buffer
 */
func (pkt *packetField) read(reader *bufio.Reader) (err os.Error) {
	var n, bytes int
	// Read catalog
	pkt.catalog, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes = n
	// Read database name
	pkt.database, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read table name
	pkt.table, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read original table name
	pkt.orgtable, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read name
	pkt.name, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read original name
	pkt.orgName, n, err = readlengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Filler 1 x 0x00
	err = readFill(reader, 1)
	if err != nil { return err }
	// Read charset
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.charsetNumber = uint16(num)
	// Read length
	num, err = readNumber(reader, 4)
	if err != nil { return err }
	pkt.length = uint32(num)
	// Read type
	pkt.fieldType, err = reader.ReadByte()
	if err != nil { return err }
	// Read flags
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.flags = uint16(num)
	// Read decimals
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.decimals = uint8(c)
	// Filler 2 x 0x00
	err = readFill(reader, 2)
	if err != nil { return err }
	bytes += 13
	// Read default if set
	if int(pkt.header.length) > bytes {
		pkt.fieldDefault, _, err = readlengthCodedBinary(reader)
		if err != nil { return err }
	}
	return
}

/**
 * Row data packet
 */
type packetRowData struct {
	header		*packetHeader
	fieldCount	uint64
	nullBitMap	byte
	values		[]string
}

/**
 * Read row data packet
 */
func (pkt *packetRowData) read(reader *bufio.Reader) (err os.Error) {
	// Read (check if exists) null bit map
	c, err := reader.ReadByte()
	if err != nil { return err }
	if c >= 0x40 && c <= 0x7f {
		pkt.nullBitMap = c
	} else {
		reader.UnreadByte()
	}
	// Read field values
	pkt.values = make([]string, pkt.fieldCount)
	for i := 0; i < int(pkt.fieldCount); i ++ {
		pkt.values[i], _, err = readlengthCodedString(reader)
		if err != nil { return err }
	}
	return nil
}

/**
 * End of file packet (received as a marker for end of packet type within a result set)
 */
type packetEOF struct {
	header		*packetHeader
	fieldCount	uint8
	warningCount	uint16
	serverStatus	uint16
}

/**
 * Read the EOF packet from the buffer
 */
func (pkt *packetEOF) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.fieldCount = uint8(c)
	// Read warning count
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.warningCount = uint16(num)
	// Read server status
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.serverStatus = uint16(num)
	return
}

/**
 * Ok packet for prepared statements
 */
type packetOKPrepared struct {
	header		*packetHeader
	fieldCount	uint8
	statementId	uint32
	columnCount	uint16
	paramCount	uint16
	warningCount	uint16
}
/**
 * Read prepared statement ok packet
 */
func (pkt *packetOKPrepared) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.fieldCount = uint8(c)
	// Read statement id
	num, err := readNumber(reader, 4)
	if err != nil { return err }
	pkt.statementId = uint32(num)
	// Read column count
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.columnCount = uint16(num)
	// Read param count
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.paramCount = uint16(num)
	// Read filler 1 x 0x00
	err = readFill(reader, 1)
	if err != nil { return err }
	// Read warning count
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.warningCount = uint16(num)
	return
}

/**
 * Parameter packet
 */
type packetParameter struct {
	header		*packetHeader
	paramType	[]byte
	flags		uint16
	decimals	uint8
	length		uint32
}

/**
 * Read parameter packet
 */
func (pkt *packetParameter) read(reader *bufio.Reader) (err os.Error) {
	// Skip this packet, documentation is incorrect and it is also ignored in MySQL code!
	bytes := make([]byte, pkt.header.length)
	reader.Read(bytes)
	return
}

/**
 * Long data packet
 */
type packetLongData struct {
	header		*packetHeader
	sequence	uint8	
	statementId	uint32
	paramNumber	uint16
	paramType	[]byte
	data		string
}

/**
 * Write long data packet
 */
func (pkt *packetLongData) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 8 + uint32(len(pkt.data))
	pkt.header.sequence = pkt.sequence
	err = pkt.header.write(writer)
	if err != nil { return err }
	// Write statement id
	err = writeNumber(writer, uint64(pkt.statementId), 4)
	if err != nil { return err }
	// Write param number
	err = writeNumber(writer, uint64(pkt.paramNumber), 2)
	if err != nil { return err }
	// Write param type
	_, err = writer.Write(pkt.paramType)
	if err != nil { return err }
	// Write data
	_, err = writer.WriteString(pkt.data)
	if err != nil { return err }
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return
}

/**
 * Execute packet
 */
type packetExecute struct {
	header		*packetHeader
	command		byte
	statementId	uint32
	flags		uint8
	iterationCount	uint32
	nullBitMap	[]byte
	newParamBound	uint8
	paramType	[][]byte
}

/**
 * Write execute packet
 */
func (pkt *packetExecute) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	pkt.header = new(packetHeader)
	pkt.header.length = 11 + uint32(len(pkt.nullBitMap)) + uint32(len(pkt.paramType) * 2)
	pkt.header.sequence = 0
	err = pkt.header.write(writer)
	// Write command
	err = writer.WriteByte(byte(pkt.command))
	if err != nil { return err }
	// Write statement id
	err = writeNumber(writer, uint64(pkt.statementId), 4)
	if err != nil { return err }
	// Write flags
	err = writer.WriteByte(byte(pkt.flags))
	if err != nil { return err }
	// Write iteration count
	err = writeNumber(writer, uint64(pkt.iterationCount), 4)
	if err != nil { return err }
	// Write null bit map
	_, err = writer.Write(pkt.nullBitMap)
	if err != nil { return err }
	// Write new parameter bound flag
	err = writer.WriteByte(byte(pkt.newParamBound))
	if err != nil { return err }
	// Write param types
	if len(pkt.paramType) > 0 {
		for _, paramType := range pkt.paramType {
			_, err = writer.Write(paramType)
			if err != nil { return err }
		}
	}
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return
}

/**
 * Read a number from the buffer that is n bytes long
 */
func readNumber(reader *bufio.Reader, n uint8) (num uint64, err os.Error) {
	p := make([]byte, n)
	_, err = reader.Read(p)
	if err != nil { return 0, err }
	num = 0
	for i := uint8(0); i < n; i ++ {
		num |= uint64(p[i]) << (i * 8)
	}
	return num, nil
}

/**
 * Read n 0x00 bytes from the buffer
 */
func readFill(reader *bufio.Reader, n int) (err os.Error) {
	p := make([]byte, n)
	_, err = reader.Read(p)
	return
}

/**
 * Read a length coded bunary number from the buffer
 */
func readlengthCodedBinary(reader *bufio.Reader) (num uint64, n int, err os.Error) {
	// Read first byte
	c, err := reader.ReadByte()
	if err != nil { return 0, 0, err }
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
			num, err = readNumber(reader, 2)
			n = 3
		// 253 following 3 = value of following 24-bit word
		case uint8(c) == 253:
			num, err = readNumber(reader, 3)
			n = 4
		// 254 following 8 = value of following 64-bit word
		case uint8(c) == 254:
			num, err = readNumber(reader, 8)
			n = 9
	}
	return
}

/**
 * Read a length coded bunary string from the buffer
 */
func readlengthCodedString(reader *bufio.Reader) (str string, n int, err os.Error) {
	// Get string length
	strlen, n, err := readlengthCodedBinary(reader)
	if err != nil { return "", 0, err }
	// Total bytes read = n + strlen
	n += int(strlen)
	// Read string
	p := make([]byte, strlen)
	_, err = reader.Read(p)
	if err != nil { return "", 0, err }
	str = string(p)
	return
}

/**
 * Write a number to the buffer using n bytes
 */
func writeNumber(writer *bufio.Writer, num uint64, n uint8) (err os.Error) {
	p := make([]byte, n)
	for i := uint8(0); i < n; i ++ {
		p[i] = byte(num >> (i * 8))
	}
	_, err = writer.Write(p)
	return
}

/**
 * Write n 0x00 bytes to the buffer
 */
func writeFill(writer *bufio.Writer, n int) (err os.Error) {
	p := make([]byte, n)
	_, err = writer.Write(p)
	return
}
