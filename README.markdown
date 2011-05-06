GoMySQL Version 0.3.2
=====================


Revision History
----------------

0.3.x series [current]

* 0.3.2 - Updated to support release.r57.1 Go build, change to usage of reflect.
* 0.3.1 - Updated to support weekly.2011-04-04 Go build, change to usage of net.Dial.
* 0.3.0 - No changes since RC2.
* 0.3.0-RC-2 - Convert additional string types (issue 47). Added a check for NULL fields in the row packet handler to prevent a crash in strconv (issue 48).
* 0.3.0-RC-1 - Fixed TestSimple unit test and added TestSimpleStatement which performs the same tests as TestSimple but uses a prepared statement throughout. Fixed and variable length strings for normal queries now return string types not []byte, text/blobs are indistinguishable so are left in []byte format which is more efficient. All integer values in prepared statements are stored as either int64 or uint64 depending on the unsigned flag, this simplifies conversion greatly when binding the result. Added ParamCount() and RowCount() methods to statements. The built in Date, Time and DateTime types can now be bound as strings in statements. Added auto-reconnect to all methods using the network and added reconnect/recovery support to Prepare and Execute functions. Statement.Reset now frees any remaining rows or complete result sets from the connection before sending the reset command so no longer requires a call to FreeResult prior to calling.
* 0.3.0-beta-1 - Added full statement and functions. Refactored packet handlers into generic functions. Added new BindResult/Fetch method to get result data from prepared statements. Added type conversions for similar types to populate the result pointers with values from the row data. Added simple type conversion to standard queries. Added automatic reconnect for a select number of operations. Added greater number of client errors from the MySQL manual. Added date/time types to allow date/time elements to be stored as integers and ints, making them more useful.
* 0.3.0-alpha-3 - Added new error structs ClientError and ServerError. Replaced majority of os.Error/os.NewError functionality with MySQL specific ClientError objects. Server error responses now return a ServerError. Removed Client.Errno and Client.Error. Added deferred error processing to reader, writer and packets to catch and errors and always return a ClientError. Rewrote auto reconnect to check for specific MySQL error codes.
* 0.3.0-alpha-2 - Added transaction wrappers, Added auto-reconnect functionality to repeatable methods. Removed mutex lock/unlocking, as it is now more appropriate that the application decides when thread safe functions are required and it's considerably safer to have a sequence such as Client.Lock(), Client.Query(...), Client.Unlock(). Added a new test which performs create, drop, select, insert and update queries on a simple demo table to test the majority of the library functionality. Added additional error messages to places where an error could be returned but there was no error number/string set. Many small changes and general improvements.
* 0.3.0-alpha-1 - First test release of new library, completely rewritten from scratch. Fully compatible with all versions of MySQL using the 4.1+ protocol and 4.0 protocol (which supports earlier versions). Fully supports old and new passwords, including old passwords using the 4.1 protocol. Includes new Go style constructors 'NewClient', 'DialTCP', 'DialUnix' replacing 'New' from the 0.2 branch. All structs have been renamed to be more user friendly, MySQL has also now been replaced with Client. Removed many dependencies on external packages such as bufio. New reader that reads the entire packet completely to a slice then processes afterwards. New writer that constructs the entire packet completely to a slice and writes in a single operation. The Client.Query function no longer returns a result set and now uses the tradition store/use result mechanism for retrieving the result and processing it's contents. The 'MultiQuery' function has been removed as this is now supported by the Client.Query function. Currently all result sets must be freed before another query can be executed either using the Result.Free() method or Client.FreeResult() method, a check for additional result sets can be made using Client.MoreResults() and the next result can be retrieved using Client.NextResult(). Client.FreeResult() is capable of reading and discarding an entire result set (provided the first result set packet has been read), a partially read result set (e.g. from Client.UseResult) or a fully stored result. Transaction support and prepared statements are NOT available in this alpha release.

0.2.x series [deprecated]

* 0.2.12 - Fix a bug in getPrepareResult() causing queries returning no fields (e.g. DROP TABLE ...) to hang.
* 0.2.11 - Skipped
* 0.2.10 - Compatibility update for Go release.2011-01-20
* 0.2.9 - Added support for MySQL 5.5
* 0.2.8 - Fixes issue #38.
* 0.2.7 - Added additional binary type support: medium int (int32/uint32), decimal (string), new decimal (string), bit ([]byte), year (uint16), set ([]byte), enum/set use string type.
* 0.2.6 - Replaced buffer checks in prepared statements, similar to change in 0.2.5, more robust method to handle end of packets.
* 0.2.5 - Fixes issue #10, removed buffer check from query function as no longer needed.
* 0.2.4 - Fixes issue #7 and related issues with prepared statement - thanks to Tom Lee [[thomaslee]](/thomaslee). New faster Escape function - thanks to [[jteeuwen]](/jteeuwen). Updated/fixed examples - thanks to [[jteeuwen]](/jteeuwen). Fixes issues (#10, #21) with reading full packet, due to some delay e.g. network lag - thanks to Michał Derkacz [[ziutek]](/ziutek) and Damian Reeves for submitting fixes for this.
* 0.2.3 - Fixes issue #6 - thanks to Tom Lee [[thomaslee]](/thomaslee).
* 0.2.2 - Resolves issue #16.
* 0.2.1 - Updated to work with latest release of Go plus 1 or 2 minor tweaks.
* 0.2.0 - Functions have been reworked and now always return os.Error to provide a generic and consistent design. Improved logging output. Improved client stability. Removed length vs buffered length checks as they don't work with packets > 4096 bytes. Added new Escape function, although this is currently only suitable for short strings. Tested library with much larger databases such as multi-gigabyte tables and multi-megabyte blogs. Many minor bug fixes. Resolved issue #3, #4 and #5.

0.1.x series [obsolete]

* 0.1.14 - Added support for long data packets.
* 0.1.13 - Added proper support for NULL bit map in binary row data packets.
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
* 0.1.2 - Added MultiQuery function to return multiple result sets as an array. [not released]
* 0.1.1 - Added support for multiple queries in a single command. [not released]
* 0.1.0 - Initial release, supporting connect, close and query functions. [not released]


About
-----

The most complete and stable MySQL client library written completely in Go. The aim of this project is to provide a library with a high level of usability, good internal error handling and to emulate similar libraries available for other languages to provide an easy migration of MySQL based systems into the Go language.

For discussions, ideas, suggestions, comments, please visit the Google Group: [https://groups.google.com/group/gomysql](https://groups.google.com/group/gomysql)

Please report bugs via the GitHub issue tracker: [https://github.com/Philio/GoMySQL/issues](https://github.com/Philio/GoMySQL/issues)


License
-------

GoMySQL 0.3 is governed by a BSD-style license that can be found in the LICENSE file.

GoMySQL 0.1.x and 0.2.x is licensed under a Creative Commons Attribution-Share Alike 2.0 UK: England & Wales License.


Compatibility
-------------

Implements the MySQL protocol version 4.0- and 4.1+

Tested on versions of MySQL 4.x, 5.x (including 5.5), MariaDB and Percona.


Thread Safety
-------------

As of version 0.3, the thread safe functionality was removed from the library, but the inherited functions from sync.Mutex were retained. The reasons for this is that the inclusions of locking/unlocking within the client itself conflicted with the new functionality that had been added and it was clear that locking should be performed within the calling program and not the library. For convenience to the programmer, the mutex functions were retained allowing for Client.Lock() and Client.Unlock() to be used for thread safe operations.

In older versions of the client from 0.1.8 - 0.2.x internal locking remains, however it is not recommended to use these versions as version 0.3.x is a much better implementation.

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

This installs the package as 'mysql' so can be imported as so:

`import "mysql"`


A note about 0.3 methods and functionality
------------------------------------------

Version 0.3 is a complete rewrite of the library and a vast improvement on the 0.2 branch, because of this there are considerable changes to available methods and usage. Please see the migration guide below for more information.


Client constants
----------------

**Client.VERSION** - The current library version.

**Client.DEFAULT_PORT** - The default port for MySQL (3306).

**Client.DEFAULT_SOCKET** - The default socket for MySQL, valid for Debian/Ubuntu systems.

**Client.MAX_PACKET_SIZE** - The maximum size of packets that will be used.

**Client.PROTOCOL_41** - Used to indicate that the 4.1+ protocol should be used to connect to the server.

**Client.PROTOCOL_40** - Used to indicate that the 4.0- protocol should be used to connect to the server.

**Client.DEFAULT_PROTOCOL** - An alias for Client.PROTOCOL_41

**Client.TCP** - Used to indicate that a TCP connection should be used.

**Client.UNIX** - Used to indicate that a unix socket connection should be used (this is faster when connecting to localhost).

**Client.LOG_SCREEN** - Send log messages to stdout.

**Client.LOG_FILE** - Send log messages to a provided file pointer.


Client properties
-----------------

**Client.LogLevel** - The level of logging to provide to stdout, valid values are 0 (none), 1 (essential information), 2 (extended information), 3 (all information).

**Client.LogType** - The type of logging to use, should be one of mysql.LOG_SCREEN or mysql.LOG_FILE, the default is mysql.LOG_SCREEN.

**Client.LogFile** - A pointer for logging to file, can be used to log to any source that implements os.File.

**Client.AffectedRows** - The number of affected rows for the last operation (if applicable).

**Client.LastInsertId** - The insert id of the last operation (if applicable).

**Client.Warnings** - The number of warnings generated by the last operation (if applicable).

**Client.Reconnect** - Set to true to enable automatic reconnect for dropped connections.


Client methods
--------------

**mysql.NewClient(protocol ...uint8) (c *Client)** - Get a new client instance, optionally specifying the protocol.

**mysql.DialTCP(raddr, user, passwd string, dbname ...string) (c *Client, err os.Error)** - Connect to the server using TCP.

**mysql.DialUnix(raddr, user, passwd string, dbname ...string) (c *Client, err os.Error)** - Connect to the server using unix socket.

**Client.Connect(network, raddr, user, passwd string, dbname ...string) (err os.Error)** - Connect to the server using the provided details.

**Client.Close() (err os.Error)** - Close the connection to the server.

**Client.ChangeDb(dbname string) (err os.Error)** - Change database.

**Client.Query(sql string) (err os.Error)** - Perform an SQL query.

**Client.StoreResult() (result *Result, err os.Error)** - Store the complete result set and return a pointer to the result.

**Client.UseResult() (result *Result, err os.Error)** - Use the result set but do not store the result, data is read from the server one row at a time via Result.Fetch functions (see below).

**Client.FreeResult() (err os.Error)** - Traditionally this function would free the memory used by the result set, in GoMySQL this removes the reference to allow the GC to clean up the memory. All results must be freed before more queries can be performed at present. FreeResult also reads and discards any remaining row packets received for the result set.

**Client.MoreResults() bool** - Check if more results are available.

**Client.NextResult() (more bool, err os.Error)** - Get the next result set from the server.

**Client.SetAutoCommit(state bool) (err os.Error)** - Set the auto commit state of the connection.

**Client.Start() (err os.Error)** - Start a new transaction.

**Client.Commit() (err os.Error)** - Commit the current transaction.

**Client.Rollback() (err os.Error)** - Rollback the current transaction.

**Client.Escape(s string) (esc string)** - Escape a string.

**Client.InitStmt() (stmt *Statement, err os.Error)** - Initialise a new statement.

**Client.Prepare(sql string) (stmt *Statement, err os.Error)** - Initialise and prepare a new statement using the supplied query.


Result methods
--------------

**Result.FieldCount() uint64** - Get the number of fields in the result set.

**Result.FetchField() *Field** - Get the next field in the result set.

**Result.FetchFields() []*Field** - Get all fields in the result set.

**Result.RowCount() uint64** - Get the number of rows in the result set, **works for stored results only**, used result always return 0.

**Result.FetchRow() Row** - Get the next row in the result set.

**Result.FetchMap() Map** - Get the next row in the result set and convert to a map with field names as keys.

**Result.FetchRows() []Row** - Get all rows in the result set, works for stored results only, used results always return nil.


Statement properties
--------------------

**Statement.AffectedRows** - The number of affected rows for the last statement operation (if applicable).

**Statement.LastInsertId** - The insert id of the last statement operation (if applicable).

**Statement.Warnings** - The number of warnings generated by the last statement operation (if applicable).


Statement methods
-----------------

**Statement.Prepare(sql string) (err os.Error)** - Prepare a new statement using the supplied query.

**Statement.ParamCount() uint16** - Get the number of parameters.

**Statement.BindParams(params ...interface{}) (err os.Error)** - Bind parameters to the statement.

**Statement.SendLongData(num int, data []byte) (err os.Error)** - Send a parameter as long data. The data can be > than the maximum packet size and will be split automatically.

**Statement.Execute() (err os.Error)** - Execute the statement.

**Statement.FieldCount() uint64** - Get the number of fields in the statement result set.

**Statement.FetchColumn() *Field** - Get the next field in the statement result set.

**Statement.FetchColumns() []*Field** - Get all fields in the statement result set.

**Statement.BindResult(params ...interface{}) (err os.Error)** - Bind the result, parameters passed to this functions should be pointers to variables which will be populated with the data from the fetched row. If a column value is not needed a nil can be used. Parameters should be of a "similar" type to the actual column value in the MySQL table, e.g. for an INT field, the parameter can be any integer type or a string and the relevant conversion is performed. Using integer sizes smaller than the size in the table is not recommended. The number of parameters bound can be equal or less than the number of fields in the table, providing more parameters than actual columns will result in a crash.

**Statement.RowCount() uint64** - Get the number of rows in the result set, **works for stored results only**, otherwise returns 0.

**Statement.Fetch() (eof bool, err os.Error)** - Fetch the next row in the result, values are populated into parameters bound using BindResult.

**Statement.StoreResult() (err os.Error)** - Store all rows for a result set,

**Statement.FreeResult() (err os.Error)** - Remove the result pointer, allowing the memory used for the result to be garbage collected.

**Statement.MoreResults() bool** - Check if more results are available.

**Statement.NextResult() (more bool, err os.Error)** - Get the next result set from the server.

**Statement.Reset() (err os.Error)** - Reset the statement.

**Statement.Close() (err os.Error)** - Close the statement.


Usage examples
--------------


1. A simple query

		// Connect to database  
		db, err := mysql.DialUnix(mysql.DEFAULT_SOCKET, "user", "password", "database")  
		if err != nil {  
			os.Exit(1)  
		}  
		// Perform query  
		err = db.Query("select * from my_table")  
		if err != nil {  
			os.Exit(1)  
		}  
		// Get result set  
		result, err := db.UseResult()  
		if err != nil {  
			os.Exit(1)  
		}  
		// Get each row from the result and perform some processing  
		for {  
			row := result.FetchRow()  
			if row == nil {  
				break  
			}  
			// ADD SOME ROW PROCESSING HERE  
		}  


2. Prepared statement


		// Define a struct to hold row data  
		type MyRow struct {  
			Id          uint64
			Name        string
			Description string
		}  

		// Connect to database  
		db, err := mysql.DialUnix(mysql.DEFAULT_SOCKET, "user", "password", "database")  
		if err != nil {  
			os.Exit(1)  
		}  
		// Prepare statement  
		stmt, err := db.Prepare("select * from my_table where name = ?")  
		if err != nil {  
			os.Exit(1)  
		}  
		// Bind params  
		err = stmt.BindParams("param")  
		if err != nil {  
			os.Exit(1)  
		}  
		// Execute statement  
		err = stmt.Execute()  
		if err != nil {  
			os.Exit(1)  
		}  
		// Define a new row to hold result
		var myrow MyRow
		// Bind result
		stmt.BindResult(&myrow.Id, &myrow.Name, &myrow.Description)
		// Get each row from the result and perform some processing  
		for {  
			eof, err := stmt.Fetch()  
			if err != nil {
				os.Exit(1)  
			}
			if eof {  
				break  
			}  
			// ADD SOME ROW PROCESSING HERE  
		}  


Auto-reconnect functionality
----------------------------

As of version 0.3.0 the library can detect network failures and try and reconnect automatically. Any methods that use the network connection support reconnect but may still return a network error (as the process is too complicated to recover) while a number of core methods are able to attempt to reconnect and recover the operation. The default setting for the feature is OFF.

Methods supporting recovery:

* Client.ChangeDb - Will attempt to reconnect and rerun the changedb command.
* Client.Query - Will attempt to reconnect and rerun the query.
* Statement.Prepare - Will attempt to reconnect and prepare the statement again.
* Statement.Execute - Will attempt to reconnect, prepare and execute the statement again. **Long data packets are not resent!**


Prepared statement notes (previously limitations)
-------------------------------------------------

This section is less relevant to the 0.3 client as it has full binary support and excellent type support but has been retained for reference.

When using prepared statements the data packets sent to/from the server are in binary format (normal queries send results as text).  

Prior to version 0.2.7 there were a number of unsupported data types in the library which limited the use of prepared statement selects to the most common field types.

As of version 0.2.7 all currently supported MySQL data types are fully supported, as well as a wide range of support of Go types for binding parameters. There are some minor limitations in the usage of unsigned numeric types, as Go does not natively support unsigned floating point numbers unsigned floats and doubles are limited to the maximum value of a signed float or double.

**Supported parameter types:**

Integer types: int, uint, int8, uint8, int16, uint16, int32, uint32, int64, uint64

Float types: float, float32, float64

Strings/other tyes: string

**Go row data formats:**

<table>
	<tr>
		<th>MySQL data type</th>
		<th>Native Go type</th>
	</tr>
	<tr>
		<td>TINYINT</td>
		<td>int8</td>
	</tr>
	<tr>
		<td>TINYINT (unsigned)</td>
		<td>uint8</td>
	</tr>
	<tr>
		<td>SMALLINT</td>
		<td>int16</td>
	</tr>
	<tr>
		<td>SMALLINT (unsigned)</td>
		<td>uint16</td>
	</tr>
	<tr>
		<td>MEDIUMINT</td>
		<td>int32</td>
	</tr>
	<tr>
		<td>MEDIUMINT (unsigned)</td>
		<td>uint32</td>
	</tr>
	<tr>
		<td>INT</td>
		<td>int32</td>
	</tr>
	<tr>
		<td>INT (unsigned)</td>
		<td>uint32</td>
	</tr>
	<tr>
		<td>BIGINT</td>
		<td>int64</td>
	</tr>
	<tr>
		<td>BIGINT (unsigned)</td>
		<td>uint64</td>
	</tr>
	<tr>
		<td>TIMESTAMP</td>
		<td>string</td>
	</tr>
	<tr>
		<td>DATE</td>
		<td>string</td>
	</tr>
	<tr>
		<td>TIME</td>
		<td>string</td>
	</tr>
	<tr>
		<td>DATETIME</td>
		<td>string</td>
	</tr>
	<tr>
		<td>YEAR</td>
		<td>string</td>
	</tr>
	<tr>
		<td>VARCHAR</td>
		<td>string</td>
	</tr>
	<tr>
		<td>TINYTEXT</td>
		<td>string</td>
	</tr>
	<tr>
		<td>MEDIUMTEXT</td>
		<td>string</td>
	</tr>
	<tr>
		<td>LONGTEXT</td>
		<td>string</td>
	</tr>
	<tr>
		<td>TEXT</td>
		<td>string</td>
	</tr>
	<tr>
		<td>TINYBLOB</td>
		<td>string</td>
	</tr>
	<tr>
		<td>MEDIUMBLOB</td>
		<td>string</td>
	</tr>
	<tr>
		<td>LONGBLOB</td>
		<td>string</td>
	</tr>
	<tr>
		<td>BLOB</td>
		<td>string</td>
	</tr>
	<tr>
		<td>BIT</td>
		<td>[]byte</td>
	</tr>
	<tr>
		<td>GEOMETRY</td>
		<td>[]byte</td>
	</tr>
</table>


Error handling
--------------

As of version 0.3.0 all functions return a ClientError or ServerError struct which contains a MySQL error code and description. The original Errno and Error public properties are deprecated.

As of version 0.2.0 all functions return os.Error. If the command succeeded the return value will be nil, otherwise it will contain the error.
If returned value is not nil then MySQL error code and description can then be retrieved from Errno and Error properties for additional info/debugging.
Prepared statements have their own copy of the Errno and Error properties.  
Generated errors attempt to follow MySQL protocol/specifications as closely as possible.


Migration guide from 0.2 - 0.3
------------------------------

1. Constructors

The original 'New()' method has gone and in it's place are a number of Go-style constructors: NewClient, DialTCP and DialUnix. This offers greater flexibility (such as the ability to connect to localhost via TCP) and simplified usage. If features such as logging are required or the 4.0 protocol, the NewClient constructor can be used to set options before connect, then Connect can be called as with previous versions.

2. Queries

As the Query method no longer returns a result set an extra step is needed here, either UseResult or StoreResult. UseResult allows you to read rows 1 at a time from the buffer which can reduce memory requirements for large result sets considerably. It is currently also a requirement to call FreeResult once you have finished with a result set, this may change with a later release. The MultiQuery method has been removed as the Query method can now support multiple queries, when using this feature you should check MoreResults and use NextResult to move to the next result set. You must free the previous result before calling NextResult, although again this may change later as we feel it would be more intuitive to free the result automatically.

3. Statements

The main changes for statements are that you must now bind the result parameters before fetching a row, this means type conversion is automated and there is no longer a need for type assertions on the rows. The bound result parameters can be individual vars or struct properties or really anything that can be passed as a pointer.
