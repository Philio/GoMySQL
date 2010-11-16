// Query example for GoMySQL
// This script will get the first 5 rows from table test1
package main

import (
	"mysql"
	"fmt"
	"os"
)

func main() {
	// Create new instance
	db := mysql.New()
	// Enable logging
	db.Logging = true
	// Connect to database
	db.Connect("localhost", "root", "********", "gotesting", "/tmp/mysql.sock")
	if db.Errno != 0 {
		fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
		os.Exit(1)
	}
	// Use UTF8
	db.Query("SET NAMES utf8");
	if db.Errno != 0 {
		fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
		os.Exit(1)
	}
	// Query database
	res := db.Query("SELECT * FROM test1 LIMIT 5")
	if db.Errno != 0 {
		fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
		os.Exit(1)
	}
	// Display results
	var row map[string] interface{}
	for {
		row = res.FetchMap()
		if row == nil {
			break
		}
		for key, value := range row {
			fmt.Printf("%s:%v\n", key, value)
		}
	}
	// Close connection
	db.Close()
}
