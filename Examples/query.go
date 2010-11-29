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
	db.Logging = false
	// Connect to database
	db.Connect("localhost", "root", "", "gotesting", "/tmp/mysql.sock")
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
	res, _ := db.Query("SELECT fname, lname FROM test1 LIMIT 5")
	if db.Errno != 0 {
		fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
		os.Exit(1)
	}
	
	defer db.Close()

	
	// Display results
	var row map[string] interface{}
	for {
		row = res.FetchMap()
		if row == nil {
			break
		}
//		for _, value := range row {
//			fmt.Printf("%v",  value)
//		}
		fmt.Printf("%s %s\n", row["fname"], row["lname"])
		

	}
	// Close connection

}
