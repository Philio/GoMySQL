GoMySQL Version 0.1.12
======================


Revision History
----------------

* 0.1.12 - Added auth struct to store authentication data. Removed logging param from New() in favour of just setting the public var. Added Reconnect() function. Bug fix in Query() causing panic for error packet responses. Added a number of examples.
* 0.1.11 - Added support for binary time fields, fixed missing zeros in time for datetime/timestamp fields.
* 0.1.10 - Changed row data to use interface{} instead of string, so rows contain data of the correct type.
* 0.1.9 - Small code tweaks, change to execute packets to allow params to contain up to 4096 bytes of data. [not released]
* 0.1.8 - Added internal mutex to make client operations thread safe.
* 0.1.7 - Added prepared statement support.
* 0.1.6 - Added Ping function.
* 0.1.5 - Clean up packet visibility all should have been private, add packet handlers for prepare/execute and related packets.
* 0.1.4 - Connect now uses ...interface{} for parameters to remove (or reduce) 'junk' required params to call the function. [not released]
* 0.1.3 - Added ChangeDb function to change the active database. [not released]
* 0.1.2 - Added MultiQuery function to return mutliple result sets as an array. [not released]
* 0.1.1 - Added support for multiple queries in a single command. [not released]
* 0.1.0 - Initial release, supporting connect, close and query functions. [not released]


To Do
-----

* Continue to add support for additional binary format options
* Add support for long data packets (long strings / blobs)
* Packets that are likely to exceed the 4096 byte write buffer need to be split up


About
-----

A MySQL client library written in Go. The aim of this project is to provide a library with a high level of usability, good interal error handling and to emulate similar libraries available for other languages to provide an easy migration of MySQL based systems into the Go language.

*GoMySQL is still in development, some features may not work as expected! This is my first project in Go, feedback/comments/suggestion are very welcome!*


Compatability
-------------

Implements the MySQL protocol version 4.1 so should work with MySQL server versions 4.1, 5.0, 5.1 and future releases.


Thread Safety
-------------

As of version 0.1.8 all client functions (including statements) should be thread safe.
At the time of writing this has been tested with the following configurations:

* GOMAXPROCS=1, 2 goroutines, 4 goroutines
* GOMAXPROCS=2, 2 goroutines, 4 goroutines


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

**MySQL.Logging** - Can be set to true or false to enable or disable logging.

**MySQL.Errno** - Error number for last operation.  
 
**MySQL.Error** - Error description for last operation.  

**MySQL.New()**

Create a new MySQL instance.

Example:

`db := mysql.New()`

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

**MySQL.Reconnect()**

Reconnect to the server using the credentials previously provided to Connect().

Example:

`connected := db.Reconnect()`

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

**MySQLResult.AffectedRows** - Number of rows affected by the query.  

**MySQLResult.InsertId** - The insert id of the row inserted by the query.  

**MySQLResult.WarningCount** - The number of warnings the server returned.  

**MySQLResult.Message** - The message returned by the server.  

**MySQLResult.Fields** - An array of fields returned by the server.  

**MySQLResult.FieldCount** - The number of fields returned by the server.  

**MySQLResult.Rows** - An array of rows returned by the server.  

**MySQLResult.RowCount** - The number of rows returned by the server.

**MySQLResult.FetchRow()**

Get the next row in the resut set.  
Returns an array or nil if there are no more rows.

Example:

`row := res.FetchRow()`

**MySQLResult.FetchMap()**

Get the next row in the resut set as a map.  
Returns a map or nil if there are no more rows.

Example:

`row := res.FetchMap()`


MySQL Statement Functions
-------------------------

**MySQLStatement.Errno** - Error number for last operation.  
 
**MySQLStatement.Error** - Error description for last operation.  

**MySQLStatement.StatementId** - The statement id.

**MySQLStatement.Params** - Array of param data [not implemented].

**MySQLStatement.ParamCount** - Number of params in the current statement.

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
Prepared statements have their own copy of the Errno and Error properties.  
Generated errors attempt to follow MySQL protocol/specifications as closely as possible.  
If a function returns a negative value (e.g. false or nil) the Errno and Error properties can be checked for details of the error.
