// Query example for GoMySQL
// This script will get the first 5 rows from table test1
package main

import (
	"mysql"
	"fmt"
	"os"
	"flag"
)

var (
	dbhost = flag.String("host", "", "Database server address.")
	dbuser = flag.String("user", "", "Database username.")
	dbpass = flag.String("pass", "", "Database password.")
	dbname = flag.String("db", "", "Database name.")
)

func main() {
	flag.Parse()

	if *dbhost == "" || *dbname == "" || *dbuser == "" {
		flag.Usage()
		os.Exit(1)
	}

	var err os.Error
	var res *mysql.MySQLResult
	var row map[string]interface{}
	var key string
	var value interface{}

	// Create new instance
	db := mysql.New()

	// Enable logging
	db.Logging = true

	// Connect to database
	if err = db.Connect(*dbhost, *dbuser, *dbpass, *dbname); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	defer db.Close()

	// Use UTF8
	if _, err = db.Query("SET NAMES utf8"); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	// Query database
	if res, err = db.Query("SELECT * FROM test1 LIMIT 5"); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}

	for {
		if row = res.FetchMap(); row == nil {
			break
		}

		for key, value = range row {
			fmt.Printf("%s:%v\n", key, value)
		}
	}

}
