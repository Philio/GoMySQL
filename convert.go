// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import "math"
import "strconv"

// bytes to int
func btoi(b []byte) int {
	return int(btoui(b))
}

// int to bytes
func itob(n int) []byte {
	return uitob(uint(n))
}

// bytes to uint
func btoui(b []byte) (n uint) {
	for i := uint8(0); i < uint8(strconv.IntSize) / 8; i++ {
		n |= uint(b[i]) << (i * 8)
	}
	return
}

// uint to bytes
func uitob(n uint) (b []byte) {
	b = make([]byte, strconv.IntSize / 8)
	for i := uint8(0); i < uint8(strconv.IntSize) / 8; i++ {
		b[i] = byte(n >> (i * 8))
	}
	return
}

// bytes to int16
func btoi16(b []byte) int16 {
	return int16(btoui16(b))
}

// int16 to bytes
func i16tob(n int16) []byte {
	return ui16tob(uint16(n))
}

// bytes to uint16
func btoui16(b []byte) (n uint16) {
	n |= uint16(b[0])
	n |= uint16(b[1]) << 8
	return
}

// uint16 to bytes
func ui16tob(n uint16) (b []byte) {
	b = make([]byte, 2)
	b[0] = byte(n)
	b[1] = byte(n >> 8)
	return
}

// bytes to int24
func btoi24(b []byte) (n int32) {
	u := btoui24(b)
	if u & 0x800000 != 0 {
		u |= 0xff000000
	}
	n = int32(u)
	return
}

// int24 to bytes
func i24tob(n int32) []byte {
	return ui24tob(uint32(n))
}

// bytes to uint24
func btoui24(b []byte) (n uint32) {
	for i := uint8(0); i < 3; i++ {
		n |= uint32(b[i]) << (i * 8)
	}
	return
}

// uint24 to bytes
func ui24tob(n uint32) (b []byte) {
	b = make([]byte, 3)
	for i := uint8(0); i < 3; i++ {
		b[i] = byte(n >> (i * 8))
	}
	return
}

// bytes to int32
func btoi32(b []byte) int32 {
	return int32(btoui32(b))
}

// int32 to bytes
func i32tob(n int32) []byte {
	return ui32tob(uint32(n))
}

// bytes to uint32
func btoui32(b []byte) (n uint32) {
	for i := uint8(0); i < 4; i++ {
		n |= uint32(b[i]) << (i * 8)
	}
	return
}

// uint32 to bytes
func ui32tob(n uint32) (b []byte) {
	b = make([]byte, 4)
	for i := uint8(0); i < 4; i++ {
		b[i] = byte(n >> (i * 8))
	}
	return
}
// bytes to int64
func btoi64(b []byte) int64 {
	return int64(btoui64(b))
}

// int64 to bytes
func i64tob(n int64) []byte {
	return ui64tob(uint64(n))
}

// bytes to uint64
func btoui64(b []byte) (n uint64) {
	for i := uint8(0); i < 8; i++ {
		n |= uint64(b[i]) << (i * 8)
	}
	return
}

// uint64 to bytes
func ui64tob(n uint64) (b []byte) {
	b = make([]byte, 8)
	for i := uint8(0); i < 8; i++ {
		b[i] = byte(n >> (i * 8))
	}
	return
}

// bytes to float32
func btof32(b []byte) float32 {
	return math.Float32frombits(btoui32(b))
}

// float32 to bytes
func f32tob(f float32) []byte {
	return ui32tob(math.Float32bits(f))
}

// bytes to float64
func btof64(b []byte) float64 {
	return math.Float64frombits(btoui64(b))
}

// float64 to bytes
func f64tob(f float64) []byte {
	return ui64tob(math.Float64bits(f))
}
