// Statement example for GoMySQL
// This script will get rows where id is between 1 and 5
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
	db.Connect("localhost", "root", "********", "gotesting")
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
	// Initialise statement
	stmt := db.InitStmt()
	// Prepare statement
	stmt.Prepare("SELECT * FROM test1 WHERE id > ? AND id < ?")
	if stmt.Errno != 0 {
		fmt.Printf("Error #%d %s\n", stmt.Errno, stmt.Error)
		os.Exit(1)
	}
	// Bind params 
	stmt.BindParams(1, 5)
	// Execute statement
	res := stmt.Execute()
	if stmt.Errno != 0 {
		fmt.Printf("Error #%d %s\n", stmt.Errno, stmt.Error)
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
	// Close statement
	stmt.Close()
	// Close connection
	db.Close()
}
