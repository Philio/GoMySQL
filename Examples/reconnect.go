// Reconnect example for GoMySQL
// This script will run forever, reconnect can be tested by restarting MySQL
// server while script is running.
package main

import (
	"mysql"
	"fmt"
	"os"
	"time"
)

// Reconnect function, attempts to reconnect once per second
func reconnect(db *mysql.MySQL, done chan int) {
	attempts := 0
	for {
		// Sleep for 1 second
		time.Sleep(1000000000)
		// Attempt to reconnect
		db.Reconnect()
		// If there was no error break for loop
		if db.Errno == 0 {
			break
		} else {
			attempts ++
			fmt.Printf("Reconnect attempt %d failed\n", attempts)
		}
	}
	done <- 1
}

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
	// Repeat query forever
	for {
		res := db.Query("select * from test1")
		// On error reconnect to the server
		if res == nil {
			fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
			done := make(chan int)
    			go reconnect(db, done)
    			<- done
		}
		// Sleep for 0.5 seconds
		time.Sleep(500000000)
	}
}
