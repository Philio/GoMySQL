// Reconnect example for GoMySQL
// This script will run forever, reconnect can be tested by restarting MySQL
// server while script is running.
package main

import (
	"mysql"
	"fmt"
	"os"
	"time"
	"flag"
)

var (
	dbhost = flag.String("host", "", "Database server address.")
	dbuser = flag.String("user", "", "Database username.")
	dbpass = flag.String("pass", "", "Database password.")
	dbname = flag.String("db", "", "Database name.")
)

// Reconnect function, attempts to reconnect once per second
func reconnect(db *mysql.MySQL, done chan bool) {
	var err os.Error
	attempts := 0

	for {
		// Sleep for 1 second
		time.Sleep(1e9)

		// Attempt to reconnect
		if err = db.Reconnect(); err != nil {
			break
		}

		attempts++
		fmt.Fprintf(os.Stderr, "Reconnect attempt %d failed\n", attempts)
	}

	done <- true
}

func main() {
	flag.Parse()

	if *dbhost == "" || *dbname == "" || *dbuser == "" {
		flag.Usage()
		os.Exit(1)
	}

	var err os.Error

	// Create new instance
	db := mysql.New()

	// Connect to database
	if err = db.Connect(*dbhost, *dbuser, *dbpass, *dbname); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

	// Ensure connection is closed on exit.
	defer db.Close()

	done := make(chan bool)

	// Repeat query forever
	for {
		if _, err = db.Query("SELECT * FROM test1 LIMIT 5"); err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			go reconnect(db, done)
			<-done
		}

		// Sleep for 0.5 seconds
		time.Sleep(5e8)
	}
}
