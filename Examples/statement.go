// Statement example for GoMySQL
// This script will get rows where id is between 1 and 5
package main

import (
	"mysql"
	"fmt"
	"os"
)

func main() {
	var err os.Error
	var res *mysql.MySQLResult
	var row map[string]interface{}
	var key string
	var value interface{}
	var stmt *mysql.MySQLStatement

	// Create new instance
	db := mysql.New()

	// Connect to database
	if err = db.Connect("localhost", "root", "********", "gotesting"); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Ensure connection is closed on exit.
	defer db.Close()

	// Use UTF8
	if _, err = db.Query("SET NAMES utf8"); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Initialise statement
	if stmt, err = db.InitStmt(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Prepare statement
	if err = stmt.Prepare("SELECT * FROM test1 WHERE id > ? AND id < ?"); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Bind params
	stmt.BindParams(1, 5)

	// Execute statement
	if res, err = stmt.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defer stmt.Close()

	// Display results
	for {
		if row = res.FetchMap(); row == nil {
			break
		}

		for key, value = range row {
			fmt.Printf("%s:%v\n", key, value)
		}
	}
}
