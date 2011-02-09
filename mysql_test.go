// GoMySQL - A MySQL client library for Go
//
// Copyright 2010-2011 Phil Bayfield. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
package mysql

import (
	"fmt"
	"os"
	"rand"
	"strconv"
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

	// Testing settings
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

	// Simple table queries
	CREATE_SIMPLE = "CREATE TABLE `simple` (`id` SERIAL NOT NULL, `number` BIGINT NOT NULL, `string` VARCHAR(32) NOT NULL, `text` TEXT NOT NULL, `datetime` DATETIME NOT NULL) ENGINE = InnoDB CHARACTER SET utf8 COLLATE utf8_unicode_ci COMMENT = 'GoMySQL Test Suite Simple Table';"
	SELECT_SIMPLE = "SELECT * FROM simple"
	INSERT_SIMPLE = "INSERT INTO simple VALUES (null, %d, '%s', '%s', NOW())"
	UPDATE_SIMPLE = "UPDATE simple SET `text` = '%s', `datetime` = NOW() WHERE id = %d"
	DROP_SIMPLE   = "DROP TABLE `simple`"

	// All types table queries
	CREATE_ALLTYPES = "CREATE TABLE `all_types` (`id` SERIAL NOT NULL, `tiny_int` TINYINT NOT NULL, `tiny_uint` TINYINT UNSIGNED NOT NULL, `small_int` SMALLINT NOT NULL, `small_uint` SMALLINT UNSIGNED NOT NULL, `medium_int` MEDIUMINT NOT NULL, `medium_uint` MEDIUMINT UNSIGNED NOT NULL, `int` INT NOT NULL, `uint` INT UNSIGNED NOT NULL, `big_int` BIGINT NOT NULL, `big_uint` BIGINT UNSIGNED NOT NULL, `decimal` DECIMAL(10,4) NOT NULL, `float` FLOAT NOT NULL, `double` DOUBLE NOT NULL, `real` REAL NOT NULL, `bit` BIT(32) NOT NULL, `boolean` BOOLEAN NOT NULL, `date` DATE NOT NULL, `datetime` DATETIME NOT NULL, `timestamp` TIMESTAMP NOT NULL, `time` TIME NOT NULL, `year` YEAR NOT NULL, `char` CHAR(32) NOT NULL, `varchar` VARCHAR(32) NOT NULL, `tiny_text` TINYTEXT NOT NULL, `text` TEXT NOT NULL, `medium_text` MEDIUMTEXT NOT NULL, `long_text` LONGTEXT NOT NULL, `binary` BINARY(32) NOT NULL, `var_binary` VARBINARY(32) NOT NULL, `tiny_blob` TINYBLOB NOT NULL, `medium_blob` MEDIUMBLOB NOT NULL, `blob` BLOB NOT NULL, `long_blob` LONGBLOB NOT NULL, `enum` ENUM('a','b','c','d','e') NOT NULL, `set` SET('a','b','c','d','e') NOT NULL, `geometry` GEOMETRY NOT NULL) ENGINE = InnoDB CHARACTER SET utf8 COLLATE utf8_unicode_ci COMMENT = 'GoMySQL Test Suite All Types Table'"
	DROP_ALLTYPES   = "DROP TABLE `all_types`"
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
		t.Logf("Error %s", err)
		t.Fail()
	}
	err = db.Close()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
}

// Test connect to server via Unix socket
func TestDialUnix(t *testing.T) {
	t.Logf("Running DialUnix test to %s", TEST_SOCK)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	err = db.Close()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
}

// Test connect to server with unprivileged database
func TestDialUnixUnpriv(t *testing.T) {
	t.Logf("Running DialUnix test to unprivileged database %s", TEST_DBNAMEUP)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAMEUP)
	if err != nil {
		t.Logf("Error %s", err)
	}
	if cErr, ok := err.(*ClientError); ok {
		if cErr.Errno != 1044 {
			t.Logf("Error #%d received, expected #1044", cErr.Errno)
			t.Fail()
		}
	}
}

// Test connect to server with nonexistant database
func TestDialUnixNonex(t *testing.T) {
	t.Logf("Running DialUnix test to nonexistant database %s", TEST_DBNAMEBAD)
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAMEBAD)
	if err != nil {
		t.Logf("Error %s", err)
	}
	if cErr, ok := err.(*ClientError); ok {
		if cErr.Errno != 1044 {
			t.Logf("Error #%d received, expected #1044", cErr.Errno)
			t.Fail()
		}
	}
}

// Test connect with bad password
func TestDialUnixBadPass(t *testing.T) {
	t.Logf("Running DialUnix test with bad password")
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_BAD_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error %s", err)
	}
	if cErr, ok := err.(*ClientError); ok {
		if cErr.Errno != 1045 {
			t.Logf("Error #%d received, expected #1045", cErr.Errno)
			t.Fail()
		}
	}
}

// Test queries on a simple table (create database, select, insert, update, drop database)
func TestSimple(t *testing.T) {
	t.Logf("Running simple table tests")
	db, err = DialUnix(TEST_SOCK, TEST_USER, TEST_PASSWD, TEST_DBNAME)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Create table")
	err = db.Query(CREATE_SIMPLE)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Insert 1000 records")
	rowMap := make(map[uint64][]string)
	for i := 0; i < 1000; i++ {
		num, str1, str2 := rand.Int(), randString(32), randString(128)
		err = db.Query(fmt.Sprintf(INSERT_SIMPLE, num, str1, str2))
		if err != nil {
			t.Logf("Error %s", err)
			t.Fail()
		}
		row := []string{fmt.Sprintf("%d", num), str1, str2}
		rowMap[db.LastInsertId] = row
	}
	t.Logf("Select inserted data")
	err = db.Query(SELECT_SIMPLE)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Use result")
	res, err := db.UseResult()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Validate inserted data")
	for {
		row := res.FetchRow()
		if row == nil {
			break
		}
		id, _ := strconv.Atoui64(row[0].(string))
		num, str1, str2 := row[1].(string), row[2].(string), row[3].(string)
		if rowMap[id][0] != num || rowMap[id][1] != str1 || rowMap[id][2] != str2 {
			t.Logf("String from database doesn't match local string")
			t.Fail()
		}
	}
	t.Logf("Free result")
	err = res.Free()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Update some records")
	for i := uint64(0); i < 1000; i += 5 {
		rowMap[i+1][2] = randString(256)
		err = db.Query(fmt.Sprintf(UPDATE_SIMPLE, rowMap[i+1][2], i+1))
		if err != nil {
			t.Logf("Error %s", err)
			t.Fail()
		}
		if db.AffectedRows != 1 {
			t.Logf("Expected 1 effected row but got %d", db.AffectedRows)
			t.Fail()
		}
	}
	t.Logf("Select updated data")
	err = db.Query(SELECT_SIMPLE)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Store result")
	res, err = db.StoreResult()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	t.Logf("Validate updated data")
	for {
		row := res.FetchRow()
		if row == nil {
			break
		}
		id, _ := strconv.Atoui64(row[0].(string))
		num, str1, str2 := row[1].(string), row[2].(string), row[3].(string)
		if rowMap[id][0] != num || rowMap[id][1] != str1 || rowMap[id][2] != str2 {
			t.Logf("%#v %#v", rowMap[id], row)
			t.Logf("String from database doesn't match local string")
			t.Fail()
		}
	}
	t.Logf("Free result")
	err = res.Free()
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}

	t.Logf("Drop table")
	err = db.Query(DROP_SIMPLE)
	if err != nil {
		t.Logf("Error %s", err)
		t.Fail()
	}
	err = db.Close()
	if err != nil {
		t.Logf("Error %s", err)
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

// Create a random string
func randString(strLen int) (randStr string) {
	strChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"
	for i := 0; i < strLen; i++ {
		randUint := rand.Uint32()
		pos := randUint % uint32(len(strChars))
		randStr += string(strChars[pos])
	}
	return
}
