package mysql

import (
	"bufio"
	"os"
	"crypto/sha1"
)

/**
 * The packet header, always at the start of every packet
 */
type PacketHeader struct {
	Length		uint32
	Sequence	uint8
}

/**
 * Read packet header from buffer
 */
func (hdr *PacketHeader) read(reader *bufio.Reader) (err os.Error) {
	// Read length
	num, err := readNumber(reader, 3)
	if err != nil { return err }
	hdr.Length = uint32(num)
	// Read sequence
	c, err := reader.ReadByte()
	if err != nil { return err }
	hdr.Sequence = uint8(c)
	return nil
}

/**
 * Write packet header to buffer
 */
func (hdr *PacketHeader) write(writer *bufio.Writer) (err os.Error) {
	// Write length
	err = writeNumber(writer, uint64(hdr.Length), 3)
	if err != nil { return err }
	// Write sequence
	err = writer.WriteByte(byte(hdr.Sequence))
	if err != nil { return err }
	return nil
}

/**
 * The initialisation packet (sent from server to client immediately following new connection
 */
type PacketInit struct {
	ProtocolVersion		uint8
	ServerVersion		string
	ThreadId		uint32
	ScrambleBuff		[]byte
	ServerCapabilities	uint16
	ServerLanguage		uint8
	ServerStatus		uint16
}

/**
 * Read initialisation packet from buffer
 */
func (pkt *PacketInit) read(reader *bufio.Reader) (err os.Error) {
	// Protocol version
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.ProtocolVersion = uint8(c)
	// Server version
	line, err := reader.ReadString(0x00)
	if err != nil { return err }
	pkt.ServerVersion = line
	// Thread id
	num, err := readNumber(reader, 4)
	if err != nil { return err }
	pkt.ThreadId = uint32(num)
	// Scramble buffer (first part)
	pkt.ScrambleBuff = new([20]byte)
	_, err = reader.Read(pkt.ScrambleBuff[0:8])
	if err != nil { return err }
	// Skip next byte
	err = readFill(reader, 1);
	if err != nil { return err }
	// Server capabilities
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.ServerCapabilities = uint16(num)
	// Server language
	c, err = reader.ReadByte()
	if err != nil { return err }
	pkt.ServerLanguage = uint8(c)
	// Server status
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.ServerStatus = uint16(num)
	// Skip next 13 bytes
	err = readFill(reader, 13);
	if err != nil { return err }
	// Scramble buffer (second part)
	_, err = reader.Read(pkt.ScrambleBuff[8:20])
	if err != nil { return err }
	// Read final byte
	err = readFill(reader, 1);
	if err != nil { return err }
	return nil
}

/**
 * The authentication packet (sent from client to server following the initialisation packet)
 */
type PacketAuth struct {
	ClientFlags	uint32
	MaxPacketSize	uint32
	CharsetNumber	uint8
	User		string
	ScrambleBuff	[]byte
	DatabaseName	string
}

/**
 * Password encryption mechanism use by MySQL = SHA1(SHA1(SHA1(password)), scramble) XOR SHA1(password)
 */
func (pkt *PacketAuth) encrypt(password string, scrambleBuff []byte) {
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
	pkt.ScrambleBuff = new([21]byte)
	pkt.ScrambleBuff[0] = 0x14
	for i := 0; i < 20; i ++ {
		pkt.ScrambleBuff[i + 1] = stg3Hash[i] ^ stg1Hash[i]
	}
}

/**
 * Write authentication packet to buffer and flush
 */
func (pkt *PacketAuth) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	hdr := new(PacketHeader)
	hdr.Length = 4 + 4 + 1 + 23 + uint32(len(pkt.User)) + 1
	if len(pkt.ScrambleBuff) > 0 {
		hdr.Length += 21
	} else {
		hdr.Length += 1
	}
	if len(pkt.DatabaseName) > 0 {
		hdr.Length += uint32(len(pkt.DatabaseName)) + 1
	}
	hdr.Sequence = 1
	err = hdr.write(writer)
	if err != nil { return err }
	// Write the client flags
	err = writeNumber(writer, uint64(pkt.ClientFlags), 4)
	if err != nil { return err }
	// Write max packet size
	err = writeNumber(writer, uint64(pkt.MaxPacketSize), 4)
	if err != nil { return err }
	// Write charset
	err = writer.WriteByte(byte(pkt.CharsetNumber))
	if err != nil { return err }
	// Filler 23 x 0x00
	err = writeFill(writer, 23)
	if err != nil { return err }
	// Write Username
	_, err = writer.WriteString(pkt.User)
	if err != nil { return err }
	// Terminate string with 0x00
	err = writeFill(writer, 1)
	if err != nil { return err }
	// Write Scramble buffer of filler 1 x 0x00
	if len(pkt.ScrambleBuff) > 0 {
		_, err = writer.Write(pkt.ScrambleBuff)
	} else {
		err = writeFill(writer, 1)
	}
	if err != nil { return err }
	// Write database name
	if len(pkt.DatabaseName) > 0 {
		_, err = writer.WriteString(pkt.DatabaseName)
		if err != nil { return err }
		// Terminate string with 0x00
		err = writeFill(writer, 1)
		if err != nil { return err }
	}
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return nil
}

/**
 * The OK packet (received after an operation completed successfully)
 */
type PacketOK struct {
	FieldCount	uint8
	AffectedRows	uint64
	InsertId	uint64
	ServerStatus	uint16
	WarningCount	uint16
	Message		string
}

/**
 * Read the OK packet from the buffer
 */
func (pkt *PacketOK) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.FieldCount = uint8(c)
	// Read affected rows
	pkt.AffectedRows, _, err = readLengthCodedBinary(reader)
	if err != nil { return err }
	// Read insert id
	pkt.InsertId, _, err = readLengthCodedBinary(reader)
	if err != nil { return err }
	// Read server status
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.ServerStatus = uint16(num)
	// Read warning count
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.WarningCount = uint16(num)
	// Read message
	if reader.Buffered() > 0 {
		msg := make([]byte, reader.Buffered())
		_, err = reader.Read(msg)
		if err != nil { return err }
		pkt.Message = string(msg)
	}
	return nil
}

/**
 * Error packet (received after an operation failed)
 */
type PacketError struct {
	FieldCount	uint8
	Errno		uint16
	State		string
	Error		string
}

/**
 * Read error packet from the buffer
 */
func (pkt *PacketError) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.FieldCount = uint8(c)
	// Read error code
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.Errno = uint16(num)
	// Read # byte
	c, err = reader.ReadByte()
	if err != nil { return err }
	// If byte isn't # then state is missing
	if c == 0x23 {
		// Read state
		state := make([]byte, 5)
		_, err = reader.Read(state)
		if err != nil { return err }
		pkt.State = string(state)
	} else {
		reader.UnreadByte()
	}
	// Read message
	msg := make([]byte, reader.Buffered())
	_, err = reader.Read(msg)
	if err != nil { return err }
	pkt.Error = string(msg)
	return nil
}

/**
 * Standard command packet (tells the server to do something defined by arg)
 */
type PacketCommand struct {
	Command	byte
	Arg	string
}

/**
 * Write the command packet to the buffer and send to the server
 */
func (pkt *PacketCommand) write(writer *bufio.Writer) (err os.Error) {
	// Construct packet header
	hdr := new(PacketHeader)
	hdr.Length = 1 + uint32(len(pkt.Arg))
	hdr.Sequence = 0
	err = hdr.write(writer)
	if err != nil { return err }
	// Write command
	err = writer.WriteByte(byte(pkt.Command))
	if err != nil { return err }
	// Write arg
	if len(pkt.Arg) > 0 {
		_, err = writer.WriteString(pkt.Arg)
		if err != nil { return err }
	}
	// Flush
	err = writer.Flush()
	if err != nil { return err }
	return nil
}

/**
 * The result set packet (used only to identify a result set)
 */
type PacketResultSet struct {
	header		*PacketHeader
	FieldCount	uint64
	Extra		uint64
}

/**
 * Read the result set from the buffer (requires the packet length as buffer will contain several packets)
 */
func (pkt *PacketResultSet) read(reader *bufio.Reader) (err os.Error) {
	var n int
	// Read field count
	pkt.FieldCount, n, err = readLengthCodedBinary(reader)
	if err != nil { return err }
	// Read extra
	if n < int(pkt.header.Length) {
		pkt.Extra, _, err = readLengthCodedBinary(reader)
		if err != nil { return err }
	}
	return nil
}

/**
 * Field packet
 */
type PacketField struct {
	header		*PacketHeader
	Catalog		string
	Database	string
	Table		string
	OrgTable	string
	Name		string
	OrgName		string
	CharsetNumber	uint16
	Length		uint32
	Type		byte
	Flags		uint16
	Decimals	uint8
	Default		uint64
}

/**
 * Read field packet from buffer
 */
func (pkt *PacketField) read(reader *bufio.Reader) (err os.Error) {
	var n, bytes int
	// Read catalog
	pkt.Catalog, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes = n
	// Read database name
	pkt.Database, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read table name
	pkt.Table, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read original table name
	pkt.OrgTable, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read name
	pkt.Name, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Read original name
	pkt.OrgName, n, err = readLengthCodedString(reader)
	if err != nil { return err }
	bytes += n
	// Filler 1 x 0x00
	err = readFill(reader, 1)
	if err != nil { return err }
	// Read charset
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.CharsetNumber = uint16(num)
	// Read length
	num, err = readNumber(reader, 4)
	if err != nil { return err }
	pkt.Length = uint32(num)
	// Read type
	pkt.Type, err = reader.ReadByte()
	if err != nil { return err }
	// Read flags
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.Flags = uint16(num)
	// Read decimals
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.Decimals = uint8(c)
	// Filler 2 x 0x00
	err = readFill(reader, 2)
	if err != nil { return err }
	bytes += 13
	// Read default if set
	if int(pkt.header.Length) > bytes {
		pkt.Default, _, err = readLengthCodedBinary(reader)
		if err != nil { return err }
	}
	return nil
}

/**
 * Row data packet
 */
type PacketRowData struct {
	header		*PacketHeader
	FieldCount	uint64
	NullBitMap	byte
	Values		[]string
}

/**
 * Read row data packet
 */
func (pkt *PacketRowData) read(reader *bufio.Reader) (err os.Error) {
	// Read (check if exists) null bit map
	c, err := reader.ReadByte()
	if err != nil { return err }
	if c >= 0x40 && c <= 0x7f {
		pkt.NullBitMap = c
	} else {
		reader.UnreadByte()
	}
	// Read field values
	pkt.Values = make([]string, pkt.FieldCount)
	for i := 0; i < int(pkt.FieldCount); i ++ {
		pkt.Values[i], _, err = readLengthCodedString(reader)
		if err != nil { return err }
	}
	return nil
}

/**
 * End of file packet (received as a marker for end of packet type within a result set)
 */
type PacketEOF struct {
	FieldCount	uint8
	WarningCount	uint16
	ServerStatus	uint16
}

/**
 * Read the EOF packet from the buffer
 */
func (pkt *PacketEOF) read(reader *bufio.Reader) (err os.Error) {
	// Read field count
	c, err := reader.ReadByte()
	if err != nil { return err }
	pkt.FieldCount = uint8(c)
	// Read warning count
	num, err := readNumber(reader, 2)
	if err != nil { return err }
	pkt.WarningCount = uint16(num)
	// Read server status
	num, err = readNumber(reader, 2)
	if err != nil { return err }
	pkt.ServerStatus = uint16(num)
	return nil
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
func readLengthCodedBinary(reader *bufio.Reader) (num uint64, n int, err os.Error) {
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
func readLengthCodedString(reader *bufio.Reader) (str string, n int, err os.Error) {
	// Get string length
	strlen, n, err := readLengthCodedBinary(reader)
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
