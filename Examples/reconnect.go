package main

import (
	"mysql"
	"fmt"
	"os"
	"time"
)

func reconnect(db *mysql.MySQL, done chan int) {
	attempts := 0
	for {
		time.Sleep(1000000000)
		db.Reconnect()
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
	db := mysql.New()
	db.Logging = true
	db.Connect("localhost", "root", ".315d*", "gotesting")
	if db.Errno != 0 {
		fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
		os.Exit(1)
	}
	// Test reconnect
	for {
		res := db.Query("select * from test1")
		if res == nil {
			fmt.Printf("Error #%d %s\n", db.Errno, db.Error)
			done := make(chan int)
    			go reconnect(db, done)
    			<-done
		}
		time.Sleep(500000000)
	}
}
