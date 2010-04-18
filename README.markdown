GoMySQL Version 0.1.7
=====================

Revision History
----------------

* 0.1.7 - Added prepared statement support
* 0.1.6 - Added Ping function
* 0.1.5 - Clean up packet visibility all should have been private, add packet handlers for prepare/execute and related packets
* 0.1.4 - Connect now uses ...interface{} for parameters to remove (or reduce) 'junk' required params to call the function
* 0.1.3 - Added ChangeDb function to change the active database
* 0.1.2 - Added MultiQuery function to return mutliple result sets as an array
* 0.1.1 - Added support for multiple queries in a single command
* 0.1.0 - Initial release, supporting connect, close and query functions


To Do
-----

* Continue to add support for additional binary format options
* Add support for long data packets (long strings / blobs)


About
-----

A MySQL client library written in Go. The aim of this project is to provide a library with a high level of usability, good interal error handling and to emulate similar libraries available for other languages to provide an easy migration of MySQL based systems into the Go language.

*GoMySQL is still in development, some features may not work as expected! This is my first project in Go, feedback/comments/suggestion are very welcome!*


Compatability
-------------

Implements the MySQL protocol version 4.1 so should work with MySQL server versions 4.1, 5.0, 5.1 and future releases.


Installation
------------

There are 2 methods to install GoMySQL

1. Via goinstall:

`goinstall github.com/Philio/GoMySQL`

The library will be installed in the same path locally so the import must be the same:

`import "github.com/Philio/GoMySQL"`

2. Via make:

Clone the git repository:

`git clone git://github.com/Philio/GoMySQL.git`

Build / install:

`cd GoMySQL`  
`make`  
`make install`

This installs the package as 'mysql' so can be importated as so:

`import "mysql"`


MySQL functions
-------------------

**MySQL.New(logging bool)**

Create a new MySQL instance, with or without logging enabled.
Logging displays all packets sent to/from the server, useful for debugging.

Example:

`db := mysql.New(false)`

**MySQL.Connect(host string, username string, [password string, [dbname string, [port int || socket string]]])**

Connect to database defined by host.  
The minimum required params to connect are host and username, dependant on server settings.  
If the host provided is localhost or 127.0.0.1 then a socket connection will be made, otherwise TCP will be used.
The fifth parameter can either be an integer value which will be assigned as the port number, or a string which will be assigned as the socket. In versions prior to 0.1.4 it was nescessary to specify both, however as both are never required together this has been changed.
The port will default to 3306.  
The socket will default to /var/run/mysql/mysql.sock (Debian/Ubuntu).  

Returns true on success or false on failure, error number and description can be retrieved for failure description (see error handling section)

Example:

`connected := db.Connect("localhost", "user", "password", "database")`

**MySQL.Close()**

Closes the connection to the database.  
Returns true on success or false on failure, error number and description can be retrieved for failure description (see error handling section)

Example:

`closed = db.Close()`

**MySQL.Query(sql string)**

Perform an SQL query, as of 0.1.1 supports multiple statements.  
Returns a MySQLResult object on success or nil on failure, data contained within result object varies depending on query type. If query contains multiple statements then the first result set is returned.

Example:

`res := db.Query("SELECT * FROM table")`

**MySQL.MultiQuery(sql string)**

Identical in function to MySQL.Query, intended for use with multiple statements.  
Returns an array of MySQLResult objects or nil on failure.

Example:

`resArray := db.MultiQuery("UPDATE t1 SET a = 1; UPDATE t2 SET b = 2")`

resArray[0] contains result of UPDATE t1 SET a = 1  
resArray[1] contains result of UPDATE t2 SET b = 2

**MySQL.ChangeDb(dbname string)**

Change the currently active database.  
Returns true on success or false on failure.

Example:  

`ok := db.ChangeDb("my_database")`

**MySQL.Ping()**

Ping the server.  
Returns true on success or false on failure.  

Example:

`ok := db.Ping()`

**MySQL.InitStmt()**

Create a new prepared statement.

Example:

`stmt := db.InitStmt()`


MySQL Result Functions
----------------------

**MySQLResult.FetchRow()**

Get the next row in the resut set.  
Returns a string array or nil if there are no more rows.

Example:

`row := res.FetchRow()`

**MySQLResult.FetchMap()**

Get the next row in the resut set as a map.  
Returns a map or nil if there are no more rows.

Example:

`row := res.FetchMap()`


MySQL Statement Functions
-------------------------

**MySQLStatement.Prepare(sql string)**

Prepare a query.  
Returns true on success or false on failure.

Example:

`ok := stmt.Prepare("SELECT a, b, c FROM table1 WHERE a > ? OR b < ?)`

**MySQLStatement.BindParams(params)**

Bind params to a query, the number of params should equal the number of ?'s in the query sent to prepare.  
Returns true on success or false on failure.  
*Please read limitations section below for supported param types*

Example:

`ok := stmt.BindParams(10, 15)`

**MySQLStatement.Execute()**

Execute the prepared query.  
Returns a MySQLResult object on success or nil on failure, data contained within result object varies depending on query type.

Example:

`res := stmt.Execute()`

**MySQLStatement.Reset()**

Reset the statement.  
Returns true on success or false on failure.

Example:

`ok := stmt.Reset()`

**MySQLStatement.Close()**

Close the statement.  
Returns true on success or false on failure.

Example:

`ok := stmt.Close()`


Prepared Statement Limitations
------------------------------

When using prepared statements the data packets sent to/from the server are in binary format (normal queries send results as text).  
Currently not all MySQL field types have been implemented.  

**Supported parameter formats:**

Format of list is Go type (MySQL type)

Integers: int (int or bigint), uint (unsigned int or big int), int8 (tiny int), uint8 (unsigned tiny int), int16 (small int), uint16 (unsigned small int), int32 (int), uint32 (unsigned int), int64 (big int), uint64 (unsigned big int)

Floats: float (float or double), float32 (float), float64 (double)

Strings: all varchar/text/blob/enum/date fields should work when sent as string

**Supported row formats:**

Format of list is MySQL type (Go type)

Integers: tiny int (int8), unsigned tiny int (uint8), small int (int16), unsigned small int (uint16), int (int32), unsigned int (uint32), big int (int64), unsigned big int (uint64)

Floats: float (float32), double (float64)

Strings: varchar, *text, *blob

Date/time: date, datetime, timestamp


Error handling
--------------

Almost all errors are handled internally and populate the Errno and Error properties of MySQL, as of 0.1.7 this includes connect errors.
Generated errors attempt to follow MySQL protocol/specifications as closely as possible.  
If a function returns a negative value (e.g. false or nil) the Errno and Error properties can be checked for details of the error.
