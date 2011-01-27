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

func TestDialTCP(t *testing.T) {
	_, err := DialTCP(TEST_HOST, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Fail()
	}
}

func TestDialUnix(t *testing.T) {
	_, err := DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Fail()
	}
}
