GoMySQL Version 0.1.3
=====================

Revision History
----------------

* 0.1   - Initial release, supporting connect, close and query functions
* 0.1.1 - Added support for multiple queries in a single command
* 0.1.2 - Added MultiQuery function to return mutliple result sets as an array
* 0.1.3 - Added ChangeDb function to change the active database


About
-----

A MySQL client library written in Go. The aim of this project is to provide a library with a high level of usability and interal error handling to allow the calling program the option to decide what to do in the event of an error and implement retry logic where required.

*GoMySQL is still in development, some features may not work as expected! This is my first project in Go, feedback/comments/suggestion are very welcome!*


Compatability
-------------

Implements the MySQL protocol version 4.1 so should work with MySQL server versions 4.1, 5.0, 5.1 and future releases.


Supported functions
-------------------

MySQL.New(logging bool)

Create a new MySQL instance, with or without logging enabled.
Logging displays all packets sent to/from the server, useful for debugging.

Example:

`db := mysql.New(false)`

MySQL.Connect(host string, username string, password string, dbname string, port int, socket string)

Connect to database defined by host.  
The minimum required params to connect are host and username, dependant on server settings.  
If the host provide is localhost or 127.0.0.1 then a socket connection will be made, otherwise TCP will be used.  
The port will default to 3306.  
The socket will default to /var/run/mysql/mysql.sock (Debian/Ubuntu).  

Returns true on success or false on failure, error number and description can be retrieved for failure description (see error handling section)

Example:

`connected := db.Connect("localhost", "user", "password", "database", 0, "")`

MySQL.Close()

Closes the connection to the database.  
Returns true on success or false on failure, error number and description can be retrieved for failure description (see error handling section)

Example:

`closed = db.Close()`

MySQL.Query(sql string)

Perform an SQL query, as of 0.1.1 supports multiple statements.  
Returns a MySQLResult object on success or nil on failure, data contained within result object varies depending on query type. If query contains multiple statements then the first result set is returned.

Example:

`res := db.Query("SELECT * FROM table")`

MySQL.MultiQuery(sql string)

Identical in function to MySQL.Query, intended for use with multiple statements.  
Returns an array of MySQLResult objects or nil on failure.

Example:

`resArray := db.MultiQuery("UPDATE t1 SET a = 1; UPDATE t2 SET b = 2")`

resArray[0] contains result of UPDATE t1 SET a = 1  
resArray[1] contains result of UPDATE t2 SET b = 2

MySQL.ChangeDb(dbname string)

Change the currently active database.  
Returns true on success of false on failure.

Example:  

`changed := db.ChangeDb("my_database")`

MySQLResult.FetchRow()

Get the next row in the resut set.  
Returns a string array or nil if there are no more rows.

Example:

`row := res.FetchRow()`

MySQLResult.FetchMap()

Get the next row in the resut set as a map.  
Returns a map or nil if there are no more rows.

Example:

`row := res.FetchMap()`


Error handling
--------------

Almost all errors are handled internally and populate the Errno and Error properties of MySQL.  
Connect errors populate ConnectErrno and ConnectError rather than Errno and Error.  
Generated errors attempt to follow MySQL protocol/specifications as closely as possible.  
Following any other operation the ConnectErrno or Errno can be checked to see if an error occured.
