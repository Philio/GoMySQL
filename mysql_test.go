// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import "testing"

const (
	// Testing credentials, run the following on server client prior to running:
	// create database gomysql_test;
	// create user gomysql_test@localhost identified by 'abc123';
	// grant all privileges on gomysql_test.* to gomysql_test@localhost;
	TEST_HOST       = "localhost"
	TEST_PORT       = 3306
	TEST_SOCK       = "/var/run/mysqld/mysqld.sock"
	TEST_USER       = "gomysql_test"
	TEST_PASSWD     = "abc123"
	TEST_BAD_PASSWD = "321cba"
	TEST_DBNAME     = "gomysql_test"
)

// Test connect to server via TCP
func TestDialTCP(t *testing.T) {
	t.Logf("Running DialTCP test to %s:%s", TEST_HOST, TEST_PORT)
	c, err := DialTCP(TEST_HOST, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error #%d: %s", c.Errno, c.Error)
		t.Fail()
	}
}

// Test connect to server via Unix socket
func TestDialUnix(t *testing.T) {
	t.Logf("Running DialUnix test to %s", TEST_SOCK)
	c, err := DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error #%d: %s", c.Errno, c.Error)
		t.Fail()
	}
}

// Benchmark connect/handshake via TCP
func BenchmarkDialTCP(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DialTCP(TEST_HOST, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	}
}

// Benchmark connect/handshake via Unix socket
func BenchmarkDialUnix(b *testing.B) {
	for i := 0; i < b.N; i++ {
		DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	}
}
