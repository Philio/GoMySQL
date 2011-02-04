// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"os"
	"testing"
)

const (
	// Testing credentials, run the following on server client prior to running:
	// create database gomysql_test;
	// create database gomysql_test2;
	// create database gomysql_test3;
	// create user gomysql_test@localhost identified by 'abc123';
	// grant all privileges on gomysql_test.* to gomysql_test@localhost;
	// grant all privileges on gomysql_test2.* to gomysql_test@localhost;
	TEST_HOST       = "localhost"
	TEST_PORT       = "3306"
	TEST_SOCK       = "/var/run/mysqld/mysqld.sock"
	TEST_USER       = "gomysql_test"
	TEST_PASSWD     = "abc123"
	TEST_BAD_PASSWD = "321cba"
	TEST_DBNAME     = "gomysql_test"  // This is the main database used for testing
	TEST_DBNAME2    = "gomysql_test2" // This is a privileged database used to test changedb etc
	TEST_DBNAMEUP   = "gomysql_test3" // This is an unprivileged database
	TEST_DBNAMEBAD  = "gomysql_bad"   // This is a nonexistant database
)

var (
	db  *Client
	err os.Error
)

// Test connect to server via TCP
func TestDialTCP(t *testing.T) {
	t.Logf("Running DialTCP test to %s:%s", TEST_HOST, TEST_PORT)
	db, err = DialTCP(TEST_HOST, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
		t.Fail()
	}
	err = db.Close()
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
		t.Fail()
	}
}

// Test connect to server via Unix socket
func TestDialUnix(t *testing.T) {
	t.Logf("Running DialUnix test to %s", TEST_SOCK)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
		t.Fail()
	}
	err = db.Close()
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
		t.Fail()
	}
}

// Test connect to server with unprivileged database
func TestDialUnixUnpriv(t *testing.T) {
	t.Logf("Running DialUnix test to unprivileged database %s", TEST_DBNAMEUP)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAMEUP)
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
	}
	if db.Errno != 1044 {
		t.Logf("Error #%d received, expected #1044", db.Errno)
		t.Fail()
	}
}

// Test connect to server with nonexistant database
func TestDialUnixNonex(t *testing.T) {
	t.Logf("Running DialUnix test to nonexistant database %s", TEST_DBNAMEBAD)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAMEBAD)
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
	}
	if db.Errno != 1044 {
		t.Logf("Error #%d received, expected #1044", db.Errno)
		t.Fail()
	}
}

// Test connect with bad password
func TestDialUnixBadPass(t *testing.T) {
	t.Logf("Running DialUnix test with bad password")
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_BAD_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error #%d: %s", db.Errno, db.Error)
	}
	if db.Errno != 1045 {
		t.Logf("Error #%d received, expected #1044", db.Errno)
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
